package download

import (
	"bytes"
	"net"

	"github.com/kmarrip/torrentz/peer"
)

// this package is for downloading the actual file
// 1. torrent file contains pieces []byte, this is always a multiple of 20
// each piece is of length pieceLength that is constant across all the pieces
// 2. the number of pices that needs to be downloaded will be t.Torrent.pieces/20
// 3. This can be done after the connection is unchoked and the sender sends an interested message
// 4. After the client is unchoked, you should 'request' for the piece
// 5. 'request' has an index,begin and length. The last two are byte offsets
