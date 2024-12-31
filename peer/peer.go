package peer

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"log"
	"net"
	"sync"
	"time"

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
	PeerId           string
	Conn             net.Conn
	BlockLength      uint32
	PeerIndex        uint32
	Choke            bool
	Bitfield         []byte
	Data             []byte
	PingTimeInterval float32
	BlockIndex       map[OffsetLengthPiece]int
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

func (p *Newpeer) New(t parse.Torrent, rIp net.IP, peerIndex uint32, port int32) {
	p.Torrent = t
	p.RemoteIp = rIp
	p.PeerIndex = peerIndex
	p.BlockLength = uint32(1 << 14)
	p.Choke = true
	p.Port = port
	p.Data = make([]byte, t.Info.PieceLength)
	p.BlockIndex = make(map[OffsetLengthPiece]int)

	for b := 0; b < int(p.Torrent.Info.PieceLength)/int(p.BlockLength); b++ {
		key := OffsetLengthPiece{Offset: uint32(b) * uint32(p.BlockLength), Length: p.BlockLength}
		p.BlockIndex[key] = 0
	}

	if p.Torrent.Info.PieceLength%int64(p.BlockLength) != 0 {
		var key OffsetLengthPiece
		key.Offset = uint32(int(p.Torrent.Info.PieceLength) / int(p.BlockLength))
		key.Length = uint32(p.Torrent.Info.PieceLength % int64(p.BlockLength))
		p.BlockIndex[key] = 0
	}
}

func (p *Newpeer) Download(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("peer handshake")
	conn, err := p.Handshake()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("peer handshake done")
	p.Conn = conn

	// the first peer message should be either Bitfield or have
	// TODO: support have peer message --> not pressing
	p.processPeerMessage()
	log.Println("Received have or Bitfield message")

	// check if the remote peer has the piece you are interested in
	// TODO: check if remote has the piece and then send the interested message --> pressing
	log.Println("Sending interested message")
	p.SendNoPayloadPeerMessage(config.Interested)

	// Unchoke the peer, so peer can download pieces from you
	log.Println("Sending unchoke message to peer")
	p.SendNoPayloadPeerMessage(config.Unchoke)

	//wait for unchoke from server
	log.Println("Waiting for Unchoke message from remote peer")
	p.processPeerMessage()

	go p.PingForPieces()

	for {
		p.processPeerMessage()
	}
}

func (p *Newpeer) PrintProgress() {
	done := 0
	for i := range p.BlockIndex {
		done += p.BlockIndex[i]
	}
	log.Printf("%d/%d blocks done\n", done, len(p.BlockIndex))
}

func (p *Newpeer) VerifyHashIntegrity() {
	log.Printf("Given hash %x\n", p.Torrent.PieceHashes[p.PeerIndex])
	log.Printf("Calculated hash %x\n", sha1.Sum(p.Data))
}

func (p *Newpeer) CheckIfPieceDone() bool {
	for i := range p.BlockIndex {
		if p.BlockIndex[i] == 0 {
			return false
		}
	}
	return true
}

func (p *Newpeer) ResetPingTimeInterval() {
	p.PingTimeInterval = 400
}

func (p *Newpeer) Close() {
	p.Conn.Close()
}

func (p *Newpeer) PingForPieces() {
	for {
		time.Sleep(time.Duration(p.PingTimeInterval) * time.Millisecond)
		if p.Choke == true {
			// ping time increases by 10% everytime, this is reset after an unchoke or piece message
			p.PingTimeInterval *= 1.1
			continue
		}
		p.SendRequestPeerMessage()
	}
}

func (p *Newpeer) SendRequestPeerMessage() {
	// block is made of pieces
	var key OffsetLengthPiece
	for i := range p.BlockIndex {
		if p.BlockIndex[i] == 0 {
			key.Offset = i.Offset
			key.Length = i.Length
			break
		}
	}
	log.Printf("Sending request message to peer, offset %d, length %d\n", key.Offset, key.Length)
	// 4-byte message length
	// 1-byte message ID
	// payload
	// 4-byte piece index
	// 4-byte block offset
	// 4-byte block length
	var buff bytes.Buffer
	messageLength := 13
	binary.Write(&buff, binary.BigEndian, int32(messageLength))
	buff.Write([]byte{6})
	binary.Write(&buff, binary.BigEndian, p.PeerIndex)
	binary.Write(&buff, binary.BigEndian, key.Offset)
	binary.Write(&buff, binary.BigEndian, key.Length)

	p.Conn.Write(buff.Bytes())
}

func (p *Newpeer) AddBlock(buff []byte) {
	// format for piece buffer Data

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

	p.BlockIndex[OffsetLengthPiece{Offset: uint32(blockIndex), Length: uint32(len(dataBuffer))}] = 1
	p.PrintProgress()
  p.VerifyHashIntegrity()
}
