package tracker

import (
	"github.com/kmarrip/torrentz/peer"
	"github.com/zeebo/bencode"
	"io"
	"log"
	"net/http"
)

func httpAnnoucer(annouceUrl string) []peer.Peer {
	resp, err := http.Get(annouceUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response TrackerResponse
	bencode.DecodeBytes(body, &response)
	return response.Peers
}
