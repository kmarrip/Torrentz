package main

import (
	"fmt"
	"github.com/kmarrip/torrentz/parse"
	"log"
	"os"
)

func main() {
	fd, err := os.Open("./debian.torrent")
	if err != nil {
		log.Fatalln("Error opening torrent file")
	}
	defer fd.Close()

	torrent := parse.ParseTorrent(fd)
	fmt.Printf("Announce part of the file %s,%d,%d\n", torrent.Info.Name, torrent.Info.Length, torrent.Info.PieceLength)
	fmt.Println(torrent.AnnounceList)
	fmt.Println(torrent.InfoHash)
}
