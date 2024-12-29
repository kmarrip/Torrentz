package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"

	"github.com/kmarrip/torrentz/parse"
)

// handshake
// once you have the list of peers, open a tcp connection
// first send 19"BitTorrent protocol"
// second send 8 reserved bytes which are always 0
// thrid send the 20 byte hash_info of the info part of the torrent file
// fourth send the 20 bytes self selected peerId that was selected first --> is this my peerId or other peer's ID ??
func VerifyPeerHandshakeResponse(peerHandeshakeResponse []byte, peer Peer, t parse.Torrent) error {
	prefixLength := peerHandeshakeResponse[:1][0]
	if prefixLength != byte(19) {
		return errors.New("Peer handshake reponse failed")
	}
	if string(peerHandeshakeResponse[1:20]) != "BitTorrent protocol" {
		return errors.New("Peer handshake response failed, protocol doesn't match")
	}

	if reflect.DeepEqual([]byte(t.InfoHash), peerHandeshakeResponse[28:48]) == false {
		return errors.New("Info hash from the peer didn't match")
	}
	return nil
}

func Handshake(t parse.Torrent, peer Peer) (net.Conn, error) {
	var buffer bytes.Buffer

	buffer.Write([]byte{19})
	buffer.WriteString("BitTorrent protocol")
	buffer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	buffer.Write([]byte(t.InfoHash))
	buffer.Write([]byte(t.PeerId))
	log.Println(string(buffer.Bytes()))
	var netDial string
	if peer.IpAddress.To4() == nil {
		// now the ip address is in ipv6 format
		netDial = fmt.Sprintf("[%s]:%d", peer.IpAddress, peer.Port)
	} else {
		netDial = fmt.Sprintf("%s:%d", peer.IpAddress, peer.Port)
	}
	log.Println(netDial)
	conn, err := net.Dial("tcp", netDial)
	if err != nil {
		log.Println("connection to remote failed")
		log.Fatal(err)
	}
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Fatalln(err)
	}
	buff := make([]byte, 68)
	// the first response sent would be
	io.ReadFull(conn, buff)
	err = VerifyPeerHandshakeResponse(buff, peer, t)
	if err != nil {
		log.Println("Peer handshake response failed, closing connection")
		return nil, err
	}
	return conn, nil
}
