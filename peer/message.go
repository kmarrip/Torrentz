package peer

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/kmarrip/torrentz/config"
)

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

func (p *Newpeer) SendNoPayloadPeerMessage(messageType int) {
	// interested, not interested, choke and unchoke are no payload messages
	var buff bytes.Buffer
	messageLength := 1
	binary.Write(&buff, binary.BigEndian, int32(messageLength))
	buff.Write([]byte{byte(messageType)})
	p.Conn.Write(buff.Bytes())
}
