package main

import (
	"log"
	"math/rand"
	"os"

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
	log.Printf("Total peers found %d\n", len(peers))

  // this is where the files will be downloaded
  os.Mkdir(torrent.Info.Name, 0755)

	newPeer := peer.Newpeer{}

	// get a random peer and see if the piece can be downloaded
	remotePeer := peers[rand.Intn(len(peers))]
	newPeer.New(torrent, remotePeer.IpAddress, int32(remotePeer.Port),0)
	newPeer.Download()
	log.Println("Done")
}
