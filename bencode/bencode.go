package bencode

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"go-torrent/peers"
	"io"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces       string `bencode:"pieces"`
	PiecesLength int    `bencode:"piece length"`
	Length       int    `bencode:"length"`
	Name         string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

func Open(r io.Reader) (*bencodeTorrent, error) {
	becodeT := bencodeTorrent{}
	err := bencode.Unmarshal(r, &becodeT)
	if err != nil {
		return nil, err
	}
	return &becodeT, nil
}

func (bt bencodeTorrent) ToTorrentFile() (TorrentFile, error) {
	infoHash, err := bt.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	pieceHashes, err := bt.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	return TorrentFile{
		Announce:    bt.Announce,
		Name:        bt.Info.Name,
		Length:      bt.Info.Length,
		PieceLength: bt.Info.PiecesLength,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
	}, nil
}

func (bi bencodeInfo) hash() ([20]byte, error) {
	// Re-encode the struct
	var bencodedInfo bytes.Buffer
	err := bencode.Marshal(&bencodedInfo, bi)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(bencodedInfo.Bytes()), nil
}

func (bi bencodeInfo) splitPieceHashes() ([]peers.Hash, error) {
	// Calculate the pieces hashes
	piecesBytes := []byte(bi.Pieces)

	// Ensure that the length of piecesBytes is a multiple of 20
	if len(piecesBytes)%20 != 0 {
		return []peers.Hash{}, fmt.Errorf("invalid piece length")
	}

	// number of pieces
	nPieces := len(piecesBytes) / 20

	pieceHashes := make([]peers.Hash, nPieces)

	for i := 0; i < nPieces; i++ {
		var hash [20]byte
		copy(hash[:], piecesBytes[i*20:(i+1)*20])
		pieceHashes[i] = hash
	}

	return pieceHashes, nil
}
