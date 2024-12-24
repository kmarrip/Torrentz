package parse

import (
	"bytes"
	"crypto/sha1"
	"io"
	"log"
	"net/url"

	"github.com/zeebo/bencode"
)

type Torrent struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	InfoHash     string
	Info         struct {
		Name        string `bencode:"name,omitempty"`
		Length      int64  `bencode:"length"`
		PieceLength int64  `bencode:"piece length"`
		Pieces      []byte `bencode:"pieces"`
	} `bencode:"info"`

	// TODO
	// multi files are not supported here
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
	t.InfoHash = url.QueryEscape(string(hashValue[:]))
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
	return torrent
}
