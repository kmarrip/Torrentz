package peer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"time"

	"github.com/kmarrip/torrentz/config"
)

// handshake
// once you have the list of peers, open a tcp connection
// first send 19"BitTorrent protocol"
// second send 8 reserved bytes which are always 0
// thrid send the 20 byte hash_info of the info part of the torrent file
// fourth send the 20 bytes self selected peerId that was selected first --> is this my peerId or other peer's ID ??
func (p *Newpeer) VerifyPeerHandshakeResponse(peerHandeshakeResponse []byte) error {
	prefixLength := peerHandeshakeResponse[:1][0]
	if prefixLength != byte(19) {
		return errors.New("Peer handshake reponse failed")
	}
	if string(peerHandeshakeResponse[1:20]) != "BitTorrent protocol" {
		return errors.New("Peer handshake response failed, protocol doesn't match")
	}

	if reflect.DeepEqual([]byte(p.Torrent.InfoHash), peerHandeshakeResponse[28:48]) == false {
		return errors.New("Info hash from the peer didn't match")
	}
	return nil
}

func (p *Newpeer) Handshake() (net.Conn, error) {
	var buffer bytes.Buffer
	buffer.Write([]byte{19})
	buffer.WriteString("BitTorrent protocol")
	buffer.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	buffer.Write([]byte(p.Torrent.InfoHash))
	buffer.Write([]byte(p.Torrent.PeerId))
	var netDial string
	if p.RemoteIp.To4() == nil {
		// now the ip address is in ipv6 format
		netDial = fmt.Sprintf("[%s]:%d", p.RemoteIp, p.Port)
	} else {
		netDial = fmt.Sprintf("%s:%d", p.RemoteIp, p.Port)
	}

	// TODO: this needs to be removed, just for debugging
	config.CopyToClipboard(p.RemoteIp.String())
	conn, err := net.DialTimeout("tcp", netDial, 2*time.Second)
	if err != nil {
		log.Println("connection to remote peer failed")
		return nil, err
	}
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Println("write content to the tcp buffer failed")
		return nil, err
	}
	buff := make([]byte, 68)
	// the first response sent would be 
	conn.Read(buff)
	err = p.VerifyPeerHandshakeResponse(buff)
	if err != nil {
		log.Println("Peer handshake response failed, closing connection")
		return nil, err
	}
	return conn, nil
}
