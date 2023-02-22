package peers

import (
    "crypto/rand"
    // "encoding/binary"
    // "errors"
    "fmt"
    "net"
)

type PeerID [20]byte

var prefix = []byte("-P00001-")

// Peer contains the details of a peer
type Peer struct {
    IP   net.IP
    Port uint16
}

// Unmarshal parses the given byte slice and returns a slice of Peer objects
func Unmarshal() ([]Peer, error) {
	peers := make([]Peer, 1)
	peer := make([]byte, 4)
	peer[0] = 192
    
	peer[1] = 168
	peer[2] = 63
	peer[3] = 30

    
	peers[0].IP = net.IP(peer)
	peers[0].Port = 8080

	
    return peers, nil
}

// String returns a string representation of the Peer object
func (p Peer) String() string {
    return net.JoinHostPort(p.IP.String(), fmt.Sprint(p.Port))
}

// GeneratePeerID generates a new peer ID
func GeneratePeerID() (PeerID, error) {
    var id PeerID
    copy(id[:], prefix)
    _, err := rand.Read(id[len(prefix):])
    if err != nil {
        return PeerID{}, err
    }
    return id, nil
}
