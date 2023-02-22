package client

import (
	"Bittorrent-client/bitfield"
	"Bittorrent-client/handshake"
	"Bittorrent-client/message"
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"
)

// A Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     Peer
	infoHash [20]byte
	peerID   [20]byte
}
type PeerID [20]byte

var prefix = []byte("-P00001-")

// Peer contains the details of a peer
type Peer struct {
    IP   net.IP
    Port uint16
}

// Unmarshal returns the seeders as an array of peers(ip:port)
func Unmarshal() ([]Peer, error) {
	peers := make([]Peer, 1)
	peer := make([]byte, 4)
	peer[0] = 127
	peer[1] = 0
	peer[2] = 0
	peer[3] = 1

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


func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("File Mismatch: expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}

func getBitField(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("Expected bitfield but got %s", msg)
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

// New connects with a peer, completes a handshake, and receives a handshake
// returns an err if any of those fail.
func New(peer Peer, peerID, infoHash [20]byte) (*Client, error) {
	log.Println(peer.String())
	conn, err := net.DialTimeout("tcp", peer.String(), 10*time.Second)
	if err != nil {
		log.Println("fails here", err)
		return nil, err
	}

	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := getBitField(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   false,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}

func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

func (c *Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}
