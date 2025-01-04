package peer

import (
	"crypto/sha1"
	"fmt"
	"os"
)

func (p *PeerConnection) WritePiece() {
	writeToFile := fmt.Sprintf("./%s/%x", p.Torrent.Info.Name, sha1.Sum(p.Data))
	os.WriteFile(writeToFile, p.Data, 0644)
}

func (p *PeerConnection) VerifyHashIntegrity() bool {
	givenHash := fmt.Sprintf("%x", p.Torrent.PieceHashes[p.PieceIndex])
	calculatedHash := fmt.Sprintf("%x", sha1.Sum(p.Data))
	return givenHash == calculatedHash
}

func (p *PeerConnection) ResetPingTimeInterval() {
	p.PingTimeInterval = 400
}

func (p *PeerConnection) CheckForPieceInRemote() bool {
	zero := 1 << 7
	has := int(p.Bitfield[int(p.PieceIndex)/8]) & (zero >> (int(p.PieceIndex) % 8))
	if has != 0 {
		return true
	}
	return false
}
