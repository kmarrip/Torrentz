package main

import (
	"log"
	"os"
	"sync"

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

	totalGoThreads := 1
	var wg sync.WaitGroup

	for i := 0; i < totalGoThreads; i++ {
		newPeer := peer.Newpeer{}
		newPeer.New(torrent, peers[i].IpAddress, 0, int32(peers[i].Port))
		wg.Add(1)
		go newPeer.Download(&wg)
	}
	wg.Wait()
	log.Println("Done")
}
