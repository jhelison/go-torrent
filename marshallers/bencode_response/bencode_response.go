package bencoderesponse

import (
	"io"

	"github.com/jackpal/bencode-go"
)

// bencodeResponse is the response from a announce
// It stores the interval and peers
// More information about the announce response can be found on:
// https://wiki.theory.org/BitTorrent_Tracker_Protocol
type bencodeResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// Unmarshal reads a io reader and convert the bytes into a bencodeResponse
func Unmarshal(r io.Reader) (*bencodeResponse, error) {
	announceRes := bencodeResponse{}
	err := bencode.Unmarshal(r, &announceRes)
	if err != nil {
		return nil, err
	}
	return &announceRes, nil
}
