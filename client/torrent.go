package client

import (
	"crypto/rand"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/jhelison/go-torrent/marshallers/bencode"
	bencoderesponse "github.com/jhelison/go-torrent/marshallers/bencode_response"
	"github.com/jhelison/go-torrent/marshallers/peer"
)

// TorrentFromTorrrentFile returns a torrent from a torrent file
func TorrentFromTorrentFile(tFile string) (Torrent, error) {
	// Read the file
	file, err := os.Open(tFile)
	if err != nil {
		return Torrent{}, err
	}
	defer file.Close()

	// Create a new torrent from the bencode
	torrent, err := bencode.Unmarshal(file)
	if err != nil {
		return Torrent{}, err
	}

	// Transform into a torrent file
	torrentFile, err := torrent.ToTorrentFile()
	if err != nil {
		return Torrent{}, err
	}

	// Generate random bytes to be used as our client ID
	var randomBytes [20]byte
	_, err = rand.Read(randomBytes[:])
	if err != nil {
		return Torrent{}, err
	}

	// Build the announce tracker URL with our default port
	url, err := torrentFile.BuildTrackerURL(randomBytes, 6881)
	if err != nil {
		return Torrent{}, err
	}

	// Get the response and unmarshal into a bencode response
	body, err := get(url)
	if err != nil {
		return Torrent{}, err
	}
	res, err := bencoderesponse.Unmarshal(body)
	if err != nil {
		return Torrent{}, err
	}

	// Parse the peers
	peers, err := peer.Unmarshal([]byte(res.Peers))
	if err != nil {
		return Torrent{}, err
	}

	return Torrent{
		Peers:       peers,
		PeerID:      randomBytes,
		InfoHash:    torrentFile.InfoHash,
		PieceHashes: torrentFile.PieceHashes,
		PieceLength: torrentFile.PieceLength,
		Length:      torrentFile.Length,
		Name:        torrentFile.Name,
	}, nil
}

// get is just a simple https getter
func get(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return http.NoBody, err
	}

	return resp.Body, nil
}
