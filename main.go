package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/kmarrip/torrentz/config"
	"github.com/kmarrip/torrentz/parse"
	"github.com/kmarrip/torrentz/peer"
	"github.com/kmarrip/torrentz/tracker"
	"github.com/schollz/progressbar/v3"
)

func main() {
	marker := os.Args[1]
	linkOrfile := os.Args[2]
	var torrent parse.Torrent

	switch marker {
	case "-m":
		mag := parse.ParseMagnetLink(linkOrfile)
		log.Println(mag)
		log.Println("Magnet link support in progress")
		os.Exit(0)

	case "-t":
		fd, err := os.Open(linkOrfile)
		if err != nil {
			log.Fatalln("Error opening torrent file")
		}
		defer fd.Close()
		torrent = parse.ParseTorrent(fd)

	}

	log.Println(torrent.Announce)
	peers := tracker.GetPeers(torrent)

	if len(peers) == 0 {
		log.Fatalln("No peers found for this torrent")
	}

	log.Printf("Total pieces %d\n", len(torrent.PieceHashes))
	log.Printf("Total peers found %d\n", len(peers))
	log.Printf("Peer Id %s\n", torrent.PeerId)

	bar := progressbar.Default(int64(len(torrent.PieceHashes)))

	// this is where the files will be downloaded
	os.Mkdir(torrent.Info.Name, 0755)

	jobs := make(chan int, len(torrent.PieceHashes))
	doneJobs := make(chan int, len(torrent.PieceHashes))

	for i := 0; i < len(torrent.PieceHashes); i++ {
		jobs <- i
	}

	for w := 0; w < config.MaxWorkers; w++ {
		go worker(jobs, doneJobs, torrent, peers, bar)
	}

	for len(doneJobs) < len(torrent.PieceHashes) {
		//until there are no more jobs to do
	}
	log.Println("All pieces downloaded, Reassembling pieces")
	torrent.ReassemblePieces()
}

func worker(jobs chan int, doneJobs chan int, torrent parse.Torrent, peers []peer.Peer, bar interface{ Add(int) error }) {
	for job := range jobs {
		// get a random peer and see if the piece can be downloaded
		//log.Printf("Picking up %d piece index\n", job)
		remotePeer := peers[rand.Intn(len(peers))]
		peerConn := peer.PeerConnection{}

		peerConn.New(torrent, remotePeer.IpAddress, int32(remotePeer.Port), uint32(job))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := peerConn.DownloadWithTimeout(ctx)

		if err != nil {
			// download failed re-enque the job
			jobs <- job
			continue
		}
		doneJobs <- job
		bar.Add(1)
	}
}
