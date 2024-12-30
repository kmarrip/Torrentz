package peer

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
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

type Newpeer struct {
	Torrent          parse.Torrent
	RemoteIp         net.IP
	Port             int32
	PeerId           string
	Conn             net.Conn
	BlockLength      uint32
	BlockIndex       uint32
	PeerIndex        uint32
	Choke            bool
	Bitfield         []byte
	Data             bytes.Buffer
	PieceHash        []byte
	PingTimeInterval float32
}

//connection to a peer happens in this way
//1. Handshake, client(this code) talks to a peer via handshake messages. if the handshake is successful then move on
//2. Bitfield, this message is sent by the upstream peer (this is optional)
//3. Wait for unchoke message from the peer
//4. Send interested message to the peer

func (p *Newpeer) New(t parse.Torrent, rIp net.IP, peerIndex uint32, port int32) {
	p.Torrent = t
	p.RemoteIp = rIp
	p.PeerIndex = peerIndex
	p.BlockLength = uint32(2 << 13)
	p.BlockIndex = uint32(0)
	p.Choke = true
	p.Port = port
	p.PingTimeInterval = 400
}

func (p *Newpeer) Download(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("peer handshake")
	conn, err := p.Handshake()
	log.Println("peer handshake done")
	p.Conn = conn
	if err != nil {
		log.Println(err)
		return
	}

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

	// Wait until you receive an unchoke message
	// you can't request for pieces until the remote has unchoked you
	//TODO: this can be written better, the connection starts out as choked
	// remove this and it should still work fine
	p.processPeerMessage()
	log.Println("Got unchoke back")

	go p.PingForPieces()

	for {
		//TODO: check if the buffer has reached its limit
		if len(p.Data.Bytes()) == int(p.Torrent.Info.PieceLength) {
			// we have the entire Piece break now
			break
		}
		p.processPeerMessage()
	}
	if sha1.Sum(p.Data.Bytes()) == [20]byte(p.Torrent.PieceHashes[p.PeerIndex]) {
		log.Println("Piece Downloaded successfully")
	}
	p.Close()
}

func (p *Newpeer) Close() {
	p.Conn.Close()
}

func (p *Newpeer) PingForPieces() {
	// wait for 100 milli seconds
	// check if the piece is fully downloaded, if yes exit
	// if not send a request message for the next buffer index
	for {
		log.Printf("%v current choke \n", p.Choke)
		time.Sleep(time.Duration(p.PingTimeInterval) * time.Millisecond)
		if len(p.Data.Bytes()) == int(p.Torrent.Info.PieceLength) {
			return
		}
		if p.Choke == true {
			log.Println("Currently choked, can't send request messages further")
			p.PingTimeInterval *= 1.1
			log.Printf("Extending time interval by 10 percent, new interval %f\n", p.PingTimeInterval)
			continue
		}
		p.SendRequestPeerMessage()
	}
}

func (p *Newpeer) AddBlock(buff []byte) {
	p.Data.Write(buff)
	p.BlockIndex += p.BlockLength
}
