package bencode

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"

	"go-torrent/marshallers/handshake"
)

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type bencodeInfo struct {
	Pieces       string `bencode:"pieces"`
	PiecesLength int    `bencode:"piece length"`
	Length       int    `bencode:"length"`
	Name         string `bencode:"name"`
}

// Unmarshal reads a stream and translates into bencode torrent
func Unmarshal(r io.Reader) (*bencodeTorrent, error) {
	becodeT := bencodeTorrent{}
	err := bencode.Unmarshal(r, &becodeT)
	if err != nil {
		return nil, err
	}
	return &becodeT, nil
}

// ToTorrentFile transforms a bencodeTorrent into a torrentFile object
func (bt bencodeTorrent) ToTorrentFile() (TorrentFile, error) {
	// Takes the hash from the info
	infoHash, err := bt.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	// Split the info into pieces
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

// hash hashes a bencodeInfo
func (bi bencodeInfo) hash() ([20]byte, error) {
	// Re-encode the struct
	var bencodedInfo bytes.Buffer
	err := bencode.Marshal(&bencodedInfo, bi)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(bencodedInfo.Bytes()), nil
}

// splitPieceHashes splits the info into hashes
// Each hash has 20 bytes
func (bi bencodeInfo) splitPieceHashes() ([]handshake.Hash, error) {
	// Calculate the pieces hashes
	piecesBytes := []byte(bi.Pieces)

	// Ensure that the length of piecesBytes is a multiple of 20
	if len(piecesBytes)%20 != 0 {
		return []handshake.Hash{}, fmt.Errorf("invalid piece length")
	}

	// number of pieces
	nPieces := len(piecesBytes) / 20
	pieceHashes := make([]handshake.Hash, nPieces)
	for i := 0; i < nPieces; i++ {
		var hash [20]byte
		copy(hash[:], piecesBytes[i*20:(i+1)*20])
		pieceHashes[i] = hash
	}

	return pieceHashes, nil
}
