package peer

import (
	"crypto/sha1"
	"fmt"
	"os"
)

func (p *Newpeer) WritePiece() {
	writeToFile := fmt.Sprintf("./%s/%x", p.Torrent.Info.Name, sha1.Sum(p.Data))
	os.WriteFile(writeToFile, p.Data, 0644)
}

func (p *Newpeer) CheckIfPieceDone() bool {
	for _, i := range p.ping.Range() {
		if p.ping.Get(i) == 0 {
			return false
		}
	}
	return true
}

func (p *Newpeer) VerifyHashIntegrity() bool {
	givenHash := fmt.Sprintf("%x", p.Torrent.PieceHashes[p.PeerIndex])
	calculatedHash := fmt.Sprintf("%x", sha1.Sum(p.Data))
	return givenHash == calculatedHash
}

func (p *Newpeer) ResetPingTimeInterval() {
	p.PingTimeInterval = 400
}

func (p *Newpeer) CheckForPieceInRemote() bool {
	zero := 1 << 7
	has := int(p.Bitfield[int(p.PeerIndex)/8]) & (zero >> (int(p.PeerIndex) % 8))
	if has != 0 {
		return true
	}
	return false
}
