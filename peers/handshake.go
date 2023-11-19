package peers

import (
	"fmt"
	"io"
)

type Hash [20]byte
type PeerID [20]byte

type Handshake struct {
	Pstr     string
	InfoHash Hash
	PeerID   PeerID
}

func NewHandshake(peerID PeerID, infoHash Hash) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serialize is the serialization for the handshake
// It's formed by:
// - The len of the protocol ID
// - The protocol ID
// - Eight reserved bytes (all turned to zero, used for extensions)
// - The infoHash
// - The peer ID
func (h Handshake) Serialize() []byte {
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

func ReadHandshake(r io.Reader) (*Handshake, error) {
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

	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	h := Handshake{
		Pstr:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
