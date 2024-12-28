package peer

import "net"

type Peer struct {
	PeerId    string `bencode:"peer id"`
	Port      int64  `bencode:"port,omitEmpty"`
	IpAddress net.IP `bencode:"ip"`
}
