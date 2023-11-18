package peers

import "io"

type InfoHash [20]byte
type PeerID [20]byte

type Handshake struct {
	Pstr     string
	InfoHash InfoHash
	PeerID   PeerID
}

func NewHandshake(peerID PeerID, infoHash InfoHash) *Handshake {
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
	// TODO: Implement me

	return nil, nil
}
