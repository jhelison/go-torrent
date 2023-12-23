package bencode_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-torrent/marshallers/bencode"
	"go-torrent/peers"
)

// TestBuildTrackerURL tests the BuildTrackerURL
func TestBuildTrackerURL(t *testing.T) {
	testCases := []struct {
		name        string
		torrentFile bencode.TorrentFile
		errContains string
	}{
		{
			name: "pass",
			torrentFile: bencode.TorrentFile{
				Announce:    "http://torrent.test.org:6969/announce",
				InfoHash:    [20]byte{},
				PieceHashes: []peers.Hash{},
				PieceLength: 0,
				Length:      0,
				Name:        "test",
			},
		},
		{
			name: "err",
			torrentFile: bencode.TorrentFile{
				Announce: "$%",
			},
			errContains: "invalid URL escape",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := tc.torrentFile.BuildTrackerURL([20]byte{}, 1)

			if tc.errContains == "" {
				require.NoError(t, err)

				require.Equal(t, "http://torrent.test.org:6969/announce?compact=1&downloaded=0&info_hash=%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00&left=0&peer_id=%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00%00&port=1&uploaded=0", url)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
