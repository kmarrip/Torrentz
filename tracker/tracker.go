package tracker

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kmarrip/torrentz/parse"
	"github.com/kmarrip/torrentz/peer"
	"github.com/zeebo/bencode"
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

func GetNetworkAddress(announceUrl string) string{
  parsedUrl, err := url.Parse(announceUrl)
  if err!=nil{
    log.Fatal(err)
  }
  netUrl := parsedUrl.Host + parsedUrl.Path
	if parsedUrl.RawQuery != "" {
		netUrl += parsedUrl.RawQuery
	}
  return netUrl 
}
func udpAnnouncer(announceUrl string) []peer.Peer {
  log.Println(GetNetworkAddress(announceUrl))
  _, err := net.Dial("udp",GetNetworkAddress(announceUrl))
  if err!= nil {
    log.Fatal(err)
  }
  return nil  
}

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

func GetPeers(announceUrl string) []peer.Peer {
	parsedUrl, err := url.Parse(announceUrl)
	if err != nil {
		log.Fatal(err)
	}
	schema := parsedUrl.Scheme	

	// TODO: the schema could be https/http/tcp/udp
	// Need to support all the schemas

	switch schema {
	case "https":
		return httpAnnoucer(announceUrl)
	case "http":
		return httpAnnoucer(announceUrl)
	default:
		return httpAnnoucer(announceUrl)
	}
}
