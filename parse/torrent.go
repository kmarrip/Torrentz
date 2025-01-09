package parse

import (
	"bytes"
	"crypto/sha1"
	"io"
	"log"

	"github.com/kmarrip/torrentz/config"
	"github.com/zeebo/bencode"
)

type Torrent struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	InfoHash     string
	PeerId       string
	Info         struct {
		Name        string `bencode:"name,omitempty"`
		Length      int64  `bencode:"length"`
		PieceLength int64  `bencode:"piece length"`
		Pieces      []byte `bencode:"pieces"`
	} `bencode:"info"`
	PieceHashes [][]byte
	// TODO
	// multi files are not supported here
}

// ParseTorrent()
// input is file descriptor and the output is bdecoded torrent struct
func ParseTorrent(file io.Reader) Torrent {
	var torrent Torrent
	err := bencode.NewDecoder(file).Decode(&torrent)
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("Error parsing torrent file")
	}
	torrent.buildInfoHash()
	torrent.buildPeerId()
	torrent.buildPieceHashes()
	return torrent
}

// buildPeerId()
// builds a random 20 byte string
func (t *Torrent) buildPeerId() {
	t.PeerId = config.Generate20ByteRandomString()
}

// buildInfoHash()
// re-bencode the Info struct into bytes
// get sha1 hash of the bytes, the hash is url escaped
func (t *Torrent) buildInfoHash() {
	var buffer bytes.Buffer
	err := bencode.NewEncoder(&buffer).Encode(t.Info)
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("Can't convert info hash into bytes")
	}
	hashValue := sha1.Sum(buffer.Bytes())
	t.InfoHash = string(hashValue[:])

	// param add method, while building the query parms escapes the string
	// no need to do that here again

	// commented out, see reason above
	//t.InfoHash = url.QueryEscape(string(hashValue[:]))
}

func (t *Torrent) buildPieceHashes() {
	// sha1 is of constant length of 20 bytes
	// each peer index would have a hash of the corresponding piece
	numberOfPieces := len(t.Info.Pieces) / 20
	t.PieceHashes = make([][]byte, numberOfPieces)
	for i := range t.PieceHashes {
		t.PieceHashes[i] = make([]byte, 20)
	}

	for index := range t.PieceHashes {
		t.PieceHashes[index] = t.Info.Pieces[20*index : 20*(index+1)]
	}
}
