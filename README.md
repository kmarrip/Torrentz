# Torrentz

**Torrentz** is an implementation of the BitTorrent protocol, as described in [BEP-3](https://www.bittorrent.org/beps/bep_0003.html). This project uses several Go packages, each handling a specific part of the protocol. Below is an overview of the key components.  
This project doesn't support seeding.

## **Packages**

### **Tracker**
The `Tracker` package handles communication with trackers, which provide a list of peers participating in the torrent swarm. This includes:  
1. Querying the tracker for peer location.  
2. Supports both **HTTP/HTTPS** and **UDP** protocols for tracker communication.

### Parse
The `Parse` package is responsible for converting .torrent files and magnet links into a struct.
This data is used by other components to manage the torrent lifecycle.

### **Peer**
The `Peer` package handles peer-to-peer communication

1. **Handshake**: Initiates a connection by performing a handshake with a remote peer.  
2. **Piece Availability**: Checks whether the remote peer has the desired piece of the file.  
3. **TCP Connection**: Establishes a raw TCP connection with the remote peer to download pieces in a pipelined fashion.  
4. **Hash Verification**: Verifies the integrity of each piece by comparing its hash with the expected value.  
5. **Retry on Failure**: If a hash mismatch occurs, the piece is redownloaded from another random peer.  
6. **Write to Disk**: Upon successful verification, the 256KB in-memory piece is flushed to disk.

When all pieces are downloaded and verified, they are combined into a single output file.
