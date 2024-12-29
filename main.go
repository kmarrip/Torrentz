package main

import (
	"log"
	"os"
	"time"

	"github.com/kmarrip/torrentz/config"
	"github.com/kmarrip/torrentz/parse"
	"github.com/kmarrip/torrentz/peer"
	"github.com/kmarrip/torrentz/tracker"
)

func main() {
  inputFile := os.Args[1]
	fd, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln("Error opening torrent file")
	}
	defer fd.Close()

	torrent := parse.ParseTorrent(fd)
	peers := tracker.GetPeers(tracker.BuildTrackerUrl(torrent))

	log.Println("Sending handshake Message")
	time.Sleep(2 * time.Second)
	conn, _ := peer.Handshake(torrent, peers[0])

	// wait for bitfield message from peer
	log.Println("Sending interested peer message")
	time.Sleep(2 * time.Second)
	peer.SendNoPayloadPeerMessage(conn, config.Interested)

	log.Println("Sending request peer message")
	time.Sleep(2 * time.Second)
	peer.SendRequestPeerMessage(conn, config.Request, int(2 << 14))

	for {
		time.Sleep(10 * time.Second)
	}
}
