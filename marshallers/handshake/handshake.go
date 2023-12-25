package handshake

import (
	"fmt"
	"io"
)

type Hash [20]byte
type PeerID [20]byte

// Each handshake if formed by:
// - Pstr: A string identifier
// - InfoHash Sha1 for the meta info
// - PeerID: Unique identifier
type Handshake struct {
	Pstr     string
	InfoHash Hash
	PeerID   PeerID
}

// NewHandshake creates a new handshake
func NewHandshake(peerID PeerID, infoHash Hash) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Marshal is the serialization for the handshake
// It's formed by:
// - The len of the protocol ID
// - The protocol ID
// - Eight reserved bytes (all turned to zero, used for extensions)
// - The infoHash
// - The peer ID
func (h Handshake) Marshal() []byte {
	buf := make([]byte, len(h.Pstr)+49)

	// The len of protocol ID
	buf[0] = byte(len(h.Pstr))
	curr := 1
	// The protocol ID
	curr += copy(buf[curr:], h.Pstr)
	// The empty 8 bytes
	curr += copy(buf[curr:], make([]byte, 8))
	// The infoHash
	curr += copy(buf[curr:], h.InfoHash[:])
	// The peer ID
	copy(buf[curr:], h.PeerID[:])

	return buf
}

// Unmarshal reads a buffer and turn it into a handshake response
// It doesn't the inverse from the marshal function
func Unmarshal(r io.Reader) (*Handshake, error) {
	// Read a single byte
	// This is the pstr length
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}

	// Read the remaining 48 bytes + the length of the pstr
	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	// Get the info hash and peerID
	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	// Create the handshake object
	h := Handshake{
		Pstr:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
