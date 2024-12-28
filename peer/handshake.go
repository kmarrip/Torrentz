package peer

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/kmarrip/torrentz/parse"
)

// handshake
// once you have the list of peers, open a tcp connection
// first send 19"BitTorrent protocol"
// second send 8 reserved bytes which are always 0
// thrid send the 20 byte hash_info of the info part of the torrent file
// fourth send the 20 bytes self selected peerId that was selected first --> is this my peerId or other peer's ID ??


func Handshake(t parse.Torrent, peer Peer) {

  var buffer bytes.Buffer

  buffer.Write([]byte{19})
  buffer.WriteString("BitTorrent protocol")
  buffer.Write([]byte{0,0,0,0,0,0,0,0})
  buffer.Write([]byte(t.InfoHash))
  buffer.Write([]byte(t.PeerId))
  log.Println(string(buffer.Bytes()))
  netDial:= fmt.Sprintf("[%s]:%d",peer.IpAddress,peer.Port) 
  log.Println(netDial)
  conn,err := net.Dial("tcp",netDial)
  defer conn.Close()
  if err!=nil{
    log.Println("connection to remote failed")
    log.Fatal(err)
  }
  _,err = conn.Write(buffer.Bytes())
  if err!= nil {
    log.Fatalln(err)
  } 
  for {

    res,err := io.ReadAll(conn)
    if err!= nil{
      log.Fatalln(err)
    }
    log.Println(string(res))
  }
}
