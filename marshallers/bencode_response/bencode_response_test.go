package bencoderesponse_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	bencoderesponse "go-torrent/marshallers/bencode_response"
)

// TestUnmarshal tests the unmarshal function from the bencode response module
func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		announceRes string
		errContains string
	}{
		{
			name:        "pass",
			announceRes: `d8:intervali900e5:peers6:94:6:peers60:e`,
		},
		{
			name:        "error",
			announceRes: `d8:intervali900e5:peers6:94:`,
			errContains: "unexpected EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate data
			reader := strings.NewReader(tc.announceRes)

			// Call the unmarshal
			response, err := bencoderesponse.Unmarshal(reader)

			if tc.errContains == "" {
				require.NoError(t, err)

				// Check the response values
				require.EqualValues(t, response.Interval, 900)
				require.Equal(t, response.Peers, "94:6:p")
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
