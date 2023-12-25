package bencode

import (
	"net/url"
	"strconv"

	"go-torrent/marshallers/handshake"
)

// TorrentFile stores the basic information to handle processing
// and download of torrents
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes []handshake.Hash
	PieceLength int
	Length      int
	Name        string
}

// BuildTrackerURL takes the TorrentFile and build the tracker URL with params
// Example can be found on https://wiki.theory.org/BitTorrent_Tracker_Protocol
// For this implementation we are always passing as no parts have been downloaded yet
func (t *TorrentFile) BuildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()

	return base.String(), nil
}
