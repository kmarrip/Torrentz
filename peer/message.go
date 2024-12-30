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

  p.Mu.Lock()
	p.Conn.Read(messageLengthBuffer)
  p.Mu.Unlock()

  log.Printf("Message length buffer %x\n",messageLengthBuffer)

	messageLength := binary.BigEndian.Uint32(messageLengthBuffer)

	if messageLength == 0 {
    log.Printf("Message length %d, exiting the process message function\n", messageLength)
		return
	}
  p.Mu.Lock()
	p.Conn.Read(messageIdBuffer)
  p.Mu.Unlock()

	messageId := uint8(messageIdBuffer[0])

	messageBuffer := make([]byte, messageLength-1)

  log.Printf("Message length %d, message id %d\nmessage buffer %s\n", messageLength,messageId,messageBuffer)
  
  p.Mu.Lock()
	p.Conn.Read(messageBuffer)
  p.Mu.Unlock()

	switch messageId {
	case config.Choke:
		p.Choke = true
	case config.Unchoke:
		p.Choke = false
    p.ResetPingTimeInterval()
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
