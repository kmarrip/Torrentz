# torrentz
This project tries to implement the Bit-Torrent protcol described here [BEP-3](https://www.bittorrent.org/beps/bep_0003.html)

The protocol is implemented by these go-packages explained below

### Tracker 
Deals with the announcer part of the protocol,
The torrent is first parsed and the announcer is queried for the list of peer locations
Two ways implemented 1) http/https and 2) udp

### Parse
Parses the .torrent and magnet links to a struct

### Peer
Peer is where the actual communication happens
- First the client does a handshake with the remote peer
- Second the client checks if the remote peers has the piece the client is interested in
- Third the client opens a raw TCP connection with the remote peer and gets the pieces in pipelined fashion
- Fourth the client verifies the integrity of the file by checking the hashes
- Fifth if the hashes dont match, the piece is redownloaded from a another new random peer
- Sixth if the hashes match, the 256KB in-memory piece is flushed to disk

At the end all the pieces are written into a single file.
