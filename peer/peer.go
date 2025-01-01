package peer

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/kmarrip/torrentz/config"
	"github.com/kmarrip/torrentz/parse"
)

type Peer struct {
	PeerId    string `bencode:"peer id"`
	Port      int64  `bencode:"port,omitEmpty"`
	IpAddress net.IP `bencode:"ip"`
}

type OffsetLengthPiece struct {
	Offset uint32
	Length uint32
}

type Newpeer struct {
	Torrent          parse.Torrent
	RemoteIp         net.IP
	Port             int32
	Conn             net.Conn
	BlockLength      uint32
	PeerIndex        uint32
	Choke            bool
	Bitfield         []byte
	Data             []byte
	PingTimeInterval float32
	//BlockIndex       map[OffsetLengthPiece]int
	ping pingMap
}

//connection to a peer happens in this way
// byte offsets and indices are precalcluated

//1. Handshake, client(this code) talks to a peer via handshake messages. if the handshake is successful then move on
//2. Bitfield, this message is sent by the upstream peer (this is optional)
//3. Wait for unchoke message from the peer
//4. Send interested message to the peer
//5. Send unchoke message to the peer, This is not mandatory but internet recommended
//6. Start a new go routine, which pings requestPeerMessage
//7. Process messages as they come by

func (p *Newpeer) New(t parse.Torrent, rIp net.IP, port int32, peerIndex uint32) {
	p.Torrent = t
	p.RemoteIp = rIp
	p.PeerIndex = peerIndex
	p.BlockLength = config.BlockLength
	p.Choke = true
	p.Port = port
	p.Data = make([]byte, t.Info.PieceLength)
	p.Bitfield = make([]byte, len(t.PieceHashes))

  p.ping = pingMap{
    BlockIndex: make(map[OffsetLengthPiece]int),
  }

	for b := 0; b < int(p.Torrent.Info.PieceLength)/int(p.BlockLength); b++ {
		key := OffsetLengthPiece{Offset: uint32(b) * uint32(p.BlockLength), Length: p.BlockLength}
		p.ping.Set(key, 0)
	}

	if p.Torrent.Info.PieceLength%int64(p.BlockLength) != 0 {
		var key OffsetLengthPiece
		key.Offset = uint32(int(p.Torrent.Info.PieceLength) / int(p.BlockLength))
		key.Length = uint32(p.Torrent.Info.PieceLength % int64(p.BlockLength))
		p.ping.Set(key, 0)
	}
}

func (p *Newpeer) Download() error {
	//log.Println("peer handshake")
	conn, err := p.Handshake()
	if err != nil {
		//log.Println(err)
		return err
	}
	//log.Println("peer handshake done")
	p.Conn = conn
	defer p.Conn.Close()

	// the first peer message should be either Bitfield or have
	// TODO: support have peer message --> not pressing
	p.processPeerMessage()
	//log.Println("Received have or Bitfield message")

	// check if the remote peer has the piece you are interested in
	if !p.CheckForPieceInRemote() {
		return errors.New("Remote doesn't have the piece")
	}

	// Sending interested message, now that peer has the piece
	//log.Println("Sending interested message")
	p.SendNoPayloadPeerMessage(config.Interested)

	// Unchoke the peer, so peer can download pieces from you
	//log.Println("Sending unchoke message to peer")
	p.SendNoPayloadPeerMessage(config.Unchoke)

	//wait for unchoke from server
	//log.Println("Waiting for Unchoke message from remote peer")
	p.processPeerMessage()

	go p.PingForPieces()

	for {
		if p.VerifyHashIntegrity() {
			p.WritePiece()
			return nil
		}
		p.processPeerMessage()
	}
}

func (p *Newpeer) AddBlock(buff []byte) {
	//4-byte pieceIndex
	//4-byte piece offset
	//rest is for data

	//pieceIndexBuffer := buff[:4]
	blockIndexBuffer := buff[4:8]
	dataBuffer := buff[8:]

	//piceIndex := int(binary.BigEndian.Uint32(pieceIndexBuffer))
	blockIndex := int(binary.BigEndian.Uint32(blockIndexBuffer))
	temp := append(p.Data[:blockIndex], dataBuffer...)
	temp = append(temp, p.Data[blockIndex+len(dataBuffer):]...)
	p.Data = temp
	p.ping.Set(OffsetLengthPiece{Offset: uint32(blockIndex), Length: uint32(len(dataBuffer))}, 1)
}
