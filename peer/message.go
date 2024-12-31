package peer

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/kmarrip/torrentz/config"
)

func (p *Newpeer) processPeerMessage() {

	messageLengthBuffer := make([]byte, 4)
	io.ReadFull(p.Conn, messageLengthBuffer)
	messageLength := binary.BigEndian.Uint32(messageLengthBuffer)

	log.Printf("Message buffer length %d\n", messageLength)
	if messageLength == 0 {
		log.Printf("KeepAlive bittorrent message received")
		return
	}

	messageIdBuffer := make([]byte, 1)
	io.ReadFull(p.Conn, messageIdBuffer)

	messageId := uint8(messageIdBuffer[0])
	if messageLength == 1 {

		switch messageId {
		case 0:
			p.Choke = true
		case 1:
			p.Choke = false
			p.ResetPingTimeInterval()
		default:
			log.Printf("MessageId received %d, interested and not interested not supported yet\n", messageId)
		}

		return
	}

	messageBuffer := make([]byte, messageLength-1)
	io.ReadFull(p.Conn, messageBuffer)
	switch messageId {
	case config.Bitfield:
		p.Bitfield = messageBuffer
	case config.Piece:
		p.AddBlock(messageBuffer)
	default:
		log.Printf("MessageId %d, The remote peer sent something thats not implemented/supported \n", messageId)
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
