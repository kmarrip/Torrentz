package parse

import (
	"io"
	"log"

	"github.com/zeebo/bencode"
)

type Torrent struct {
	Announce string `bencode:"announce"`
  AnnounceList [][]string `bencode:"announce-list"`
	Info     struct {
    Name string `bencode:"name,omitempty"`
    Length int64`bencode:"length"`
    PieceLength int64 `bencode:"piece length"`
    Pieces string `bencode:"pieces"`
	} `bencode:"info"`
  // TODO
  // multi files are not supported here
}

// input is a bencoding string, the output is decoded torrent struct
func ParseTorrent(file io.Reader) Torrent {
	var torrent Torrent
	err := bencode.NewDecoder(file).Decode(&torrent)
	if err != nil {
    log.Fatalln(err)
		log.Fatalln("Error parsing torrent file")
	}
	return torrent
}
