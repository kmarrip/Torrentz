package tracker

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"net/url"

	"github.com/kmarrip/torrentz/config"
	"github.com/kmarrip/torrentz/parse"
	"github.com/kmarrip/torrentz/peer"
)

func sendUdpHandshake(conn *net.UDPConn) {

	// Offset  Size            Name            Value
	// 0       64-bit integer  protocol_id     0x41727101980 // magic constant
	// 8       32-bit integer  action          0 // connect
	// 12      32-bit integer  transaction_id
	log.Println("sending udp handshake")

	var buff bytes.Buffer
	buff.Write(config.Uint64ToBytes(uint64(0x41727101980)))
	buff.Write(config.Uint32ToBytes(uint32(0)))
	buff.Write(config.Uint32ToBytes(uint32(0)))
	conn.Write(buff.Bytes())
}

func getPeerFrom6bytes(buff []byte) peer.Peer {
	if len(buff) != 6 {
		log.Fatal("IP + port needs to be 6 bytes")
	}
	ipAddress := net.IP(buff[:4])
	port := binary.BigEndian.Uint16(buff[4:])
	return peer.Peer{IpAddress: ipAddress, Port: int64(port), PeerId: config.Generate20ByteRandomString()}
}

func processAnnounceResponse(conn *net.UDPConn) []peer.Peer {

	//Offset      Size            Name            Value
	//0           32-bit integer  action          1 // announce
	//4           32-bit integer  transaction_id
	//8           32-bit integer  interval
	//12          32-bit integer  leechers
	//16          32-bit integer  seeders
	//20 + 6 * n  32-bit integer  IP address
	//24 + 6 * n  16-bit integer  TCP port
	//20 + 6 * N
	responseBuffer := make([]byte, 1024)
	conn.Read(responseBuffer)

	action := config.FourBytesToInt32(responseBuffer[0:4])
	if action != 1 {
		log.Fatalf("Expected 1 for the action value receied %d\n", action)
	}
	//transcationId := config.FourBytesToInt32(responseBuffer[4:8])
	//interval := config.FourBytesToInt32(responseBuffer[8:12])
	leechers := config.FourBytesToInt32(responseBuffer[12:16])
	seeders := config.FourBytesToInt32(responseBuffer[16:20])
	totalPeers := leechers + seeders
	var peersList []peer.Peer
	for i := 0; i < int(totalPeers); i++ {
		peersList = append(peersList, getPeerFrom6bytes(responseBuffer[20+6*i:26+6*i]))
	}
	return peersList
}

func processConnectResponse(conn *net.UDPConn) uint64 {

	//Offset  Size            Name            Value
	//0       32-bit integer  action          0 // connect
	//4       32-bit integer  transaction_id
	//8       64-bit integer  connection_id
	reponseBuffer := make([]byte, 1024)
	conn.Read(reponseBuffer)
	action := binary.BigEndian.Uint32(reponseBuffer[:4])

	if action != 0 {
		log.Fatalf("Expected 0 action value, received %d\n", action)
	}

	//transcationBuffer := reponseBuffer[4:8]
	connectionIdBuffer := reponseBuffer[8:16]

	// I'm not checking transcationId
	//transcationId := binary.BigEndian.Uint32(transcationBuffer)
	connectionId := binary.BigEndian.Uint64(connectionIdBuffer)

	return connectionId
}

func sendUdpRequest(conn *net.UDPConn, connectionId int, infoHash []byte, peerId []byte) {

	//Offset  Size    Name    Value
	//0       64-bit integer  connection_id
	//8       32-bit integer  action          1 // announce
	//12      32-bit integer  transaction_id
	//16      20-byte string  info_hash
	//36      20-byte string  peer_id
	//56      64-bit integer  downloaded
	//64      64-bit integer  left
	//72      64-bit integer  uploaded
	//80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
	//84      32-bit integer  IP address      0 // default
	//88      32-bit integer  key
	//92      32-bit integer  num_want        -1 // default
	//96      16-bit integer  port

	var buff bytes.Buffer
	buff.Write(config.Uint64ToBytes(uint64(connectionId))) // 64 bits for connectionId
	buff.Write(config.Uint32ToBytes(uint32(1)))            // 32 bits for action
	buff.Write(config.Uint32ToBytes(uint32(0)))            // 32 bits for transcationId
	buff.Write(infoHash[:])                                // 20 byte string for infoHash
	buff.Write(peerId[:])                                  // 20 byte for peer id
	buff.Write(config.Uint64ToBytes(0))                    // 64 bits for downloaded
	buff.Write(config.Uint64ToBytes(1 << 14))              // 64 bits for left
	buff.Write(config.Uint32ToBytes(0))                    // event default is 0
	buff.Write([]byte(conn.LocalAddr().String()))          // default ip address
	buff.Write(config.Uint32ToBytes(0))                    // default for key
	buff.Write(config.Int32ToBytes(-1))                    // default for num_want
	buff.Write(config.Uint16ToBytes(5000))                 // default port, not going to use it ??
	conn.Write(buff.Bytes())
}

func udpAnnouncer(t parse.Torrent) []peer.Peer {
	log.Println(GetNetworkAddress(t.Announce))
	url, err := url.Parse(t.Announce)
	if err != nil {
		log.Fatal(err)
	}

	sourceAddr, _ := net.ResolveUDPAddr("udp", "localhost")
	destAddr, _ := net.ResolveUDPAddr("udp", url.Host)
	if destAddr.IP.To4() == nil {
		log.Fatalln("Peer IPv6 addresses are not supported yet")
	}
	conn, err := net.DialUDP("udp", sourceAddr, destAddr)

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()


  // TODO
  // There is no retry mechanism, if the server fails to respond
  // the next read blocks

	sendUdpHandshake(conn)
	connectionId := processConnectResponse(conn)
	sendUdpRequest(conn, int(connectionId), []byte(t.InfoHash), []byte(t.PeerId))
  peersList := processAnnounceResponse(conn)
	return peersList
}
