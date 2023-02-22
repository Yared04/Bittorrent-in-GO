package handshake

import (
	"fmt"
	"io"
)

// A Handshake is a special message that a peer uses to identify itself
type Handshake struct {
	ProtocolName string
	InfoHash [20]byte
	PeerID   [20]byte
}

// Connect creates a Connect handshake with the standard pstr
func Connect(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		ProtocolName:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serialize serializes the handshake to a buffer
func (handshake *Handshake) Serialize() []byte {
	buf := make([]byte, len(handshake.ProtocolName)+49)

	buf[0] = byte(len(handshake.ProtocolName))
	curr := 1
	curr += copy(buf[curr:], handshake.ProtocolName)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], handshake.InfoHash[:])
	curr += copy(buf[curr:], handshake.PeerID[:])
	return buf
}

// Read parses a handshake from a stream
func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	ProtocolNamelen := int(lengthBuf[0])

	if ProtocolNamelen == 0 {
		err := fmt.Errorf("ProtocolNamelen cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, 48+ProtocolNamelen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[ProtocolNamelen+8:ProtocolNamelen+8+20])
	copy(peerID[:], handshakeBuf[ProtocolNamelen+8+20:])

	h := Handshake{
		ProtocolName:     string(handshakeBuf[0:ProtocolNamelen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
