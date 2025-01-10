package tracker

import (
	"log"
	"net/url"
	"strconv"
	"sync"

	"github.com/kmarrip/torrentz/parse"
	"github.com/kmarrip/torrentz/peer"
)

type TrackerResponse struct {
	Complete   int64       `bencode:"complete,omitEmpty"`
	Incomplete int64       `bencode:"incomplete,omitEmpty"`
	Interval   int64       `bencode:"interval,omitEmpty"`
	Peers      []peer.Peer `bencode:"peers,omitEmpty"`
}

// BuildTrackerParams()
// TrackerUrl should be built using the
// announce url, infohash, peer_id, ip(optional), port (6881), uploaded
// downloaded, left, events (optional)

func BuildTrackerUrl(t parse.Torrent) string {
	parsedUrl, _ := url.Parse(t.Announce)
	params := url.Values{}
	params.Add("info_hash", t.InfoHash)
	params.Add("peer_id", t.PeerId)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.Itoa(int(t.Info.Length)))
	parsedUrl.RawQuery = params.Encode()
	return parsedUrl.String()
}

func GetNetworkAddress(announceUrl string) string {
	parsedUrl, err := url.Parse(announceUrl)
	if err != nil {
		log.Fatal(err)
	}
	netUrl := parsedUrl.Host + parsedUrl.Path
	if parsedUrl.RawQuery != "" {
		netUrl += parsedUrl.RawQuery
	}
	return netUrl
}

func GetPeers(t parse.Torrent) []peer.Peer {
	announceUrl := BuildTrackerUrl(t)
	parsedUrl, err := url.Parse(announceUrl)
	if err != nil {
		log.Fatal(err)
	}
	schema := parsedUrl.Scheme

	returnValue := []peer.Peer{}

	switch schema {
	case "https":
		returnValue = append(returnValue, httpAnnoucer(announceUrl)...)
	case "http":
		returnValue = append(returnValue, httpAnnoucer(announceUrl)...)
	case "udp":
		returnValue = append(returnValue, udpAnnouncer(t)...)
	default:
		log.Fatalln("Announcer schema not supported")
		return nil
	}
  var wg sync.WaitGroup
  for _,peer := range returnValue {
    wg.Add(1)
    p.BootPeer(wg)
  }
  wg.Wait()
  return returnValue 
}
