package peer

import (
	"bytes"
	"encoding/binary"
	"net"
)

type Peer struct {
	PeerId    string `bencode:"peer id"`
	Port      int64  `bencode:"port,omitEmpty"`
	IpAddress net.IP `bencode:"ip"`
}

type PeerMessage struct {
	MessageLength int32
	MessageType   int8
	DataBuffer    []byte
}

func (b *PeerMessage) New(response []byte) {
	// All the integers are big endian numbers, but TCP would convert the numbers into network byte order and send the back.
	// bitfield message would be 1 less than what is shown by the messagelength
	b.MessageLength = int32(binary.BigEndian.Uint32(response[:4]))
	b.MessageType = int8(binary.BigEndian.Uint16(response[4:5]))
	b.DataBuffer = response[5 : b.MessageLength-1]
}

//connection to a peer happens in this way
//1. Handshake, client(this code) talks to a peer via handshake messages. if the handshake is successful then move on
//2. Bitfield, this message is sent by the upstream peer (this is optional)
//3. Wait for unchoke message from the peer
//4. Send interested message to the peer
//5.

func SendNoPayloadPeerMessage(conn net.Conn, messageType int) {
	var buff bytes.Buffer
	messageLength := 1
	binary.Write(&buff, binary.BigEndian, int32(messageLength))
	buff.Write([]byte{byte(messageType)})
	conn.Write(buff.Bytes())
}

func SendRequestPeerMessage(conn net.Conn, messageType int, blockLength int) {
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
	binary.Write(&buff, binary.BigEndian, int32(0))
	binary.Write(&buff, binary.BigEndian, int32(0))
	binary.Write(&buff, binary.BigEndian, int32(blockLength))
	conn.Write(buff.Bytes())
}
