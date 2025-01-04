package peer

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"
)

// ping.go
// A seperate go routing will be spawned off to keep pinging the remote peer
// If the the client (we) are choked the pinging stops and the ping internval increases by 10% everytime
// BlockIndex is wrapper around at mutex to prevent concurrent reads/writes
type pingMap struct {
	BlockIndex map[OffsetLengthPiece]int
	mu         sync.Mutex
}

func (pm *pingMap) Set(key OffsetLengthPiece, val int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.BlockIndex[key] = val
}

func (pm *pingMap) Get(key OffsetLengthPiece) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.BlockIndex[key]
}

func (pm *pingMap) Range() []OffsetLengthPiece {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	returnValue := []OffsetLengthPiece{}
	for k := range pm.BlockIndex {
		returnValue = append(returnValue, k)
	}
	return returnValue
}

func (p *PeerConnection) PingForPieces() {
	for {
		time.Sleep(time.Duration(p.PingTimeInterval) * time.Millisecond)
		if p.Choke == true {
			// ping time increases by 10% everytime, this is reset after an unchoke or piece message
			p.PingTimeInterval *= 1.1
			continue
		}
		// This checks if there's anything to be requested for
		var key OffsetLengthPiece
		needToPing := false

		for _, i := range p.ping.Range() {
			if p.ping.Get(i) == 0 {
				key.Offset = i.Offset
				key.Length = i.Length
				needToPing = true
        p.SendRequestPeerMessage(key)
			}
		}
		if !needToPing {
			return
		}
	}
}

func (p *PeerConnection) SendRequestPeerMessage(key OffsetLengthPiece) {
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
	binary.Write(&buff, binary.BigEndian, p.PieceIndex)
	binary.Write(&buff, binary.BigEndian, key.Offset)
	binary.Write(&buff, binary.BigEndian, key.Length)

	p.Conn.Write(buff.Bytes())
}
