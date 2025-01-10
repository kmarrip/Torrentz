package peer

import (
	"context"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/kmarrip/torrentz/config"
	"github.com/kmarrip/torrentz/parse"
)

type Peer struct {
	Port      int64  `bencode:"port,omitEmpty"`
	IpAddress net.IP `bencode:"ip"`
}

type OffsetLengthPiece struct {
	Offset uint32
	Length uint32
}

type PeerConnection struct {
	Torrent          parse.Torrent
	RemoteIp         net.IP
	Port             int32
	Conn             net.Conn
	BlockLength      uint32
	PieceIndex       uint32
	Choke            bool
	Bitfield         []byte
	Data             []byte
	PingTimeInterval float32
	ping             pingMap
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

func (p *PeerConnection) New(t parse.Torrent, rIp net.IP, port int32, pieceIndex uint32) {
	p.Torrent = t
	p.RemoteIp = rIp
	p.PieceIndex = pieceIndex
	p.BlockLength = config.BlockLength
	p.Choke = true
	p.PingTimeInterval = 400
	p.Bitfield = make([]byte, len(t.PieceHashes))
	p.Port = port
	p.ping = pingMap{
		BlockIndex: make(map[OffsetLengthPiece]int),
	}

	sizeOfThisPiece := p.Torrent.Info.PieceLength
	if int(pieceIndex)+1 == len(p.Torrent.PieceHashes) {
		//The last piece may not be of full size
		sizeOfThisPiece = p.Torrent.Info.Length - (int64(pieceIndex) * p.Torrent.Info.PieceLength)
	}
	p.Data = make([]byte, sizeOfThisPiece)
	numberOfBlocks := int(sizeOfThisPiece / int64(p.BlockLength))

	for b := 0; b < numberOfBlocks; b++ {
		key := OffsetLengthPiece{Offset: uint32(b) * uint32(p.BlockLength), Length: p.BlockLength}
		p.ping.Set(key, 0)
	}

	if sizeOfThisPiece%int64(p.BlockLength) != 0 {
		var key OffsetLengthPiece
		key.Offset = uint32(numberOfBlocks) * p.BlockLength
		key.Length = uint32(sizeOfThisPiece) % p.BlockLength
		p.ping.Set(key, 0)
	}
}

func (p *PeerConnection) DownloadWithTimeout(ctx context.Context) error {
	routineChannel := make(chan int)
	go p.Download(routineChannel)
	for {
		select {
		case <-ctx.Done():
			return errors.New("Remote peer timeout")
		case val := <-routineChannel:
			if val == 0 {
				return nil
			} else {
				return errors.New("Remote connection failed")
			}
		}
	}
}

func (p *PeerConnection) BootPeer(wg sync.WaitGroup) error{
  defer wg.Done()
  conn, err := p.Handshake()
  if err != nil {
    log.Println("Peer handshake failed")
    return err
  }
  p.Conn = conn 
  // this peer connection never really gets closed

  //First message is have or Bitfield
  p.processPeerMessage()
  

  //Second send interested no matter what
  p.SendNoPayloadPeerMessage(config.Interested)

  //Third Unchoke the peer
  p.SendNoPayloadPeerMessage(config.Unchoke)

  //Fourth wait for unchoke message from peer
  p.processPeerMessage()
}

func (p *PeerConnection) Download(routineChannel chan int) {
	conn, err := p.Handshake()
	if err != nil {
		routineChannel <- 1
		return
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
		routineChannel <- 1
		return
		// return errors.New("Remote doesn't have the piece")
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
			break
		}
		p.processPeerMessage()
	}
	routineChannel <- 0
}

func (p *PeerConnection) AddBlock(buff []byte) {
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
