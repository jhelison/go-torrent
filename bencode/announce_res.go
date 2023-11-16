package bencode

import (
	"io"

	"github.com/jackpal/bencode-go"
)

type becodeResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func Unmarshal(r io.Reader) (*becodeResponse, error) {
	announceRes := becodeResponse{}
	err := bencode.Unmarshal(r, &announceRes)
	if err != nil {
		return nil, err
	}
	return &announceRes, nil
}
