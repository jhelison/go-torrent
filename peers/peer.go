package peers

import (
	"encoding/binary"
	"fmt"
	"net"
)

// This represents a single peer in a announce response
type Peer struct {
	IP   net.IP
	Port uint16
}

// Unmarshal reads the announce response and parse the peers
func Unmarshal(peersBytes []byte) ([]Peer, error) {
	// Ensure that we have the correct length to parse
	// each peer is 6 bytes
	if len(peersBytes)%6 != 0 {
		return []Peer{}, fmt.Errorf("invalid piers length")
	}

	// number of peers
	nPeers := len(peersBytes) / 6

	peers := make([]Peer, nPeers)
	for i := 0; i < nPeers; i++ {
		offset := i * 6
		// First four bytes
		peers[i].IP = net.IP(peersBytes[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peersBytes[offset+4 : offset+6])
	}

	return peers, nil
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%v", p.IP.String(), p.Port)
}
