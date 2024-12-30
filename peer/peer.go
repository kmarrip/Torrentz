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
	Torrent     parse.Torrent
	RemoteIp    net.IP
	Port        int32
	PeerId      string
	Conn        net.Conn
	BlockLength uint32
	BlockIndex  uint32
	PeerIndex   uint32
	Choke       bool
	Bitfield    []byte
	Data        bytes.Buffer
	PieceHash   []byte
  PingTimeInterval float32 
}

//connection to a peer happens in this way
//1. Handshake, client(this code) talks to a peer via handshake messages. if the handshake is successful then move on
//2. Bitfield, this message is sent by the upstream peer (this is optional)
//3. Wait for unchoke message from the peer
//4. Send interested message to the peer
//5.

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
    if len(p.Data.Bytes()) == int(p.Torrent.Info.PieceLength){
      // we have the entire Piece break now 
      break
    }
    p.processPeerMessage()
  }
  if sha1.Sum(p.Data.Bytes()) == [20]byte(p.Torrent.PieceHashes[p.PeerIndex]){
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
    log.Printf("%v current choke \n",p.Choke)
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

func (p *Newpeer) SendRequestPeerMessage() {
  log.Println("Sending request message to peer")
	// block is made of pieces

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
	binary.Write(&buff, binary.BigEndian, p.BlockIndex)
	binary.Write(&buff, binary.BigEndian, p.BlockLength)
	p.Conn.Write(buff.Bytes())
}

func (p *Newpeer) AddBlock(buff []byte) {
	p.Data.Write(buff)
	p.BlockIndex += p.BlockLength
}

func (p *Newpeer) processPeerMessage() {
	messageLengthBuffer := make([]byte, 4)
	messageIdBuffer := make([]byte, 1)

	p.Conn.Read(messageLengthBuffer)
	messageLength := binary.BigEndian.Uint32(messageLengthBuffer)
	if messageLength == 0 {
		return
	}
	p.Conn.Read(messageIdBuffer)
	messageId := uint8(messageIdBuffer[0])

  log.Printf("Processing peer message, message id %d\n", messageId)
	messageBuffer := make([]byte, messageLength-1)
	p.Conn.Read(messageBuffer)

	switch messageId {
	case config.Choke:
		p.Choke = true
	case config.Unchoke:
		p.Choke = false
	case config.Bitfield:
		p.Bitfield = messageBuffer
	case config.Piece:
		p.AddBlock(messageBuffer)
	default:
		log.Println("The remote peer sent something thats not implemented/supported")
	}
}

func (p *Newpeer) SendNoPayloadPeerMessage(messageType int) {
	// interested, not interested, choke and unchoke are no payload messages
	var buff bytes.Buffer
	messageLength := 1
	binary.Write(&buff, binary.BigEndian, int32(messageLength))
	buff.Write([]byte{byte(messageType)})
	p.Conn.Write(buff.Bytes())
}

// handshake
// once you have the list of peers, open a tcp connection
// first send 19"BitTorrent protocol"
// second send 8 reserved bytes which are always 0
// thrid send the 20 byte hash_info of the info part of the torrent file
// fourth send the 20 bytes self selected peerId that was selected first --> is this my peerId or other peer's ID ??
func (p *Newpeer) VerifyPeerHandshakeResponse(peerHandeshakeResponse []byte) error {
	prefixLength := peerHandeshakeResponse[:1][0]
	if prefixLength != byte(19) {
		return errors.New("Peer handshake reponse failed")
	}
	if string(peerHandeshakeResponse[1:20]) != "BitTorrent protocol" {
		return errors.New("Peer handshake response failed, protocol doesn't match")
	}

	if reflect.DeepEqual([]byte(p.Torrent.InfoHash), peerHandeshakeResponse[28:48]) == false {
		return errors.New("Info hash from the peer didn't match")
	}
	return nil
}

func (p *Newpeer) Handshake() (net.Conn, error) {
	var buffer bytes.Buffer
	buffer.Write([]byte{19})
	buffer.WriteString("BitTorrent protocol")
	buffer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	buffer.Write([]byte(p.Torrent.InfoHash))
	buffer.Write([]byte(p.Torrent.PeerId))
	var netDial string
	if p.RemoteIp.To4() == nil {
		// now the ip address is in ipv6 format
		netDial = fmt.Sprintf("[%s]:%d", p.RemoteIp, p.Port)
	} else {
		netDial = fmt.Sprintf("%s:%d", p.RemoteIp, p.Port)
	}

	// TODO: this needs to be removed, just for debugging
	config.CopyToClipboard(p.RemoteIp.String())
	conn, err := net.DialTimeout("tcp", netDial, 2*time.Second)
	if err != nil {
		log.Println("connection to remote peer failed")
		return nil, err
	}
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Println("write content to the tcp buffer failed")
		return nil, err
	}
	buff := make([]byte, 68)
	// the first response sent would be
  conn.Read(buff)
	err = p.VerifyPeerHandshakeResponse(buff)
	if err != nil {
		log.Println("Peer handshake response failed, closing connection")
		return nil, err
	}
	return conn, nil
}
