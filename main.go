package main

import (
	"log"
	"math/rand"
	"os"

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
	log.Printf("Total peers found %d\n", len(peers))

	// this is where the files will be downloaded
	os.Mkdir(torrent.Info.Name, 0755)

	jobs := make(chan int, len(torrent.PieceHashes))
	for i := 0; i < len(torrent.PieceHashes); i++ {
		jobs <- i
	}

  for w:= 1; w< config.MaxWorkers; w++{
    go worker(jobs,torrent,peers)
  }

  for;len(jobs)!=0;{
    // until there are no more jobs to do 
  }

}

func worker(jobs chan int, torrent parse.Torrent, peers []peer.Peer) {

	for job := range jobs {
		newPeer := peer.Newpeer{}
		// get a random peer and see if the piece can be downloaded
    //log.Printf("Picking up %d piece index\n", job)
		remotePeer := peers[rand.Intn(len(peers))]
		newPeer.New(torrent, remotePeer.IpAddress, int32(remotePeer.Port), uint32(job))
    err := newPeer.Download()
    if err != nil {
      // download failed re-enque the job
      //log.Printf("%d piece index job failed, redoing it\n",job)
      jobs <- job
    }
	}
}
