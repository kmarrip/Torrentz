package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	//"github.com/kmarrip/torrentz/config"
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
	log.Println(torrent.Announce)
	peers := tracker.GetPeers(torrent)
	log.Println(peers)
	log.Printf("Total peers found %d\n", len(peers))
	log.Printf("Total pieces %d\n", len(torrent.PieceHashes))

	// this is where the files will be downloaded
	os.Mkdir(torrent.Info.Name, 0755)

	jobs := make(chan int, len(torrent.PieceHashes))
	//doneJobs := make(chan int, len(torrent.PieceHashes))

	//for i := 0; i < len(torrent.PieceHashes); i++ {
	//	jobs <- i
	//}
	jobs <- 500

	//for w := 0; w < config.MaxWorkers; w++ {
	//	go worker(jobs, doneJobs, torrent, peers)
	//}

	//for len(doneJobs) < 1 {
		// until there are no more jobs to do
	//}
	log.Println("All pieces downloaded, Reassembling pieces")
	torrent.ReassemblePieces()
}

func worker(jobs chan int, doneJobs chan int, torrent parse.Torrent, peers []peer.Peer) {
	for job := range jobs {
		newPeer := peer.Newpeer{}
		// get a random peer and see if the piece can be downloaded
		//log.Printf("Picking up %d piece index\n", job)
		remotePeer := peers[rand.Intn(len(peers))]
		newPeer.New(torrent, remotePeer.IpAddress, int32(remotePeer.Port), uint32(job))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := newPeer.DownloadWithTimeout(ctx)
		if err != nil {
			// download failed re-enque the job
			jobs <- job
			continue
		}
		doneJobs <- job
		totalJobs := len(torrent.PieceHashes)
		log.Printf("Progress %d/%d pieces\n", len(doneJobs), totalJobs)
	}
}
