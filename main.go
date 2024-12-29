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
	log.Printf("Total peers found %d\n", len(peers))

	//1. send and receive handshake message
	log.Println("Sending handshake Message")
	time.Sleep(1 * time.Second)
	conn, err := peer.Handshake(torrent, peers[0])
	if err != nil {
		log.Print(err)
		log.Fatalln("Handshake with peer failed")
	}

	//2. wait for bitfield or have messages from the peer
  log.Println("waiting for have or bitfield message")
	time.Sleep(1 * time.Second)
  err = peer.RecieveBitfieldOrHaveMessage(conn)
  if err != nil {
    log.Print(err)
    log.Fatalln("Expected bitfield or have message from peer")
  }
  
  //3. send interested message 
  log.Println("Sending interested peer message")
  time.Sleep(1 * time.Second)
	peer.SendNoPayloadPeerMessage(conn, config.Interested)

  //4. wait for unchoke message
  log.Println("waiting for unchoke message")
  time.Sleep(1 * time.Second)
  err = peer.RecieveUnchokeMessage(conn)
  if err!= nil {
    log.Println(err)
    log.Fatalln("Issue with unchoke message")
  }
  
  //5. send requested message  
	log.Println("Sending request peer message")
	time.Sleep(1 * time.Second)
	peer.SendRequestPeerMessage(conn, config.Request, int(2<<14))

	for {
		time.Sleep(10 * time.Second)
	}
}
