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
	peer[0] = 127
	peer[1] = 0
	peer[2] = 0
	peer[3] = 1

	peers[0].IP = net.IP(peer)
	peers[0].Port = 8080

	
    // const peerSize = 6
	// // fmt.Println(peersBin, "###########################################################3")
    // numPeers := len(peersBin) / peerSize

    // if len(peersBin)%peerSize != 0 {
    //     return nil, errors.New("failed to unmarshal")
    // }

    // peers := make([]Peer, numPeers)
    // for i := 0; i < numPeers; i++ {
    //     offset := i * peerSize
    //     peers[i].IP = net.IP(peersBin[offset : offset+4])
    //     peers[i].Port = binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))
	// 	// fmt.Println(peers[i].IP, peers[i].Port, "#######################################################3")

    // }
	// // fmt.Println(peers, "#######################################################3")

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
