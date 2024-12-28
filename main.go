package main

import (
	"log"
	"os"

	"github.com/kmarrip/torrentz/parse"
	"github.com/kmarrip/torrentz/peer"
	"github.com/kmarrip/torrentz/tracker"
)

func main() {
	fd, err := os.Open("./debian.torrent")
	if err != nil {
		log.Fatalln("Error opening torrent file")
	}
	defer fd.Close()

	torrent := parse.ParseTorrent(fd)
	peers := tracker.GetPeers(tracker.BuildTrackerUrl(torrent))
	
  peer.Handshake(torrent,peers[2])
}
