package config

const (
	Choke = iota
	Unchoke
	Interested
	Notinterested
	Have
	Bitfield
	Request
	Piece
	Cancel
)

const BlockLength uint32 = 1 << 14
const MaxWorkers = 100
const UdpLocalPort = "54321"
