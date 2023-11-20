package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"go-torrent/client"
	"go-torrent/message"
	"go-torrent/peers"
	"log"
	"runtime"
	"time"
)

var (
	DownloadDeadline = 30 * time.Second
	MaxBacklog       = 5
	MaxBlockSize     = 16384
)

type Torrent struct {
	Peers       []peers.Peer
	PeerID      peers.PeerID
	InfoHash    peers.Hash
	PieceHashes []peers.Hash
	PieceLenght int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   peers.Hash
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func downloadPiece(client *client.Client, work *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  work.index,
		client: client,
		buf:    make([]byte, work.length),
	}

	// Set a deadline to skip stucked peers
	err := client.Conn.SetDeadline(time.Now().Add(DownloadDeadline))
	if err != nil {
		return nil, err
	}
	// We can ignore the error on the defer
	defer client.Conn.SetDeadline(time.Time{}) //nolint:errcheck

	for state.downloaded < work.length {
		// If unchoked download requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < work.length {
				blockSize := MaxBlockSize
				// Last block may be shorter
				if work.length-state.requested < blockSize {
					blockSize = work.length - state.requested
				}

				err := client.SendRequest(work.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}

				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) error {
	client, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		return err
	}
	defer client.Conn.Close()
	log.Printf("Handshake complete with peer %s\n", peer)

	err = client.SendUnchoke()
	if err != nil {
		return err
	}
	err = client.SendInterested()
	if err != nil {
		return err
	}

	for work := range workQueue {
		if !client.Bitfield.HasPiece(work.index) {
			workQueue <- work
			continue
		}

		buf, err := downloadPiece(client, work)
		if err != nil {
			log.Println("error downloading piece", err)
			workQueue <- work
			continue
		}

		err = checkWorkHash(work, buf)
		if err != nil {
			log.Println("integrity validation failed", work.index)
		}

		err = client.SendHave(work.index)
		if err != nil {
			log.Println("sending have failed", err)
		}
		results <- &pieceResult{
			index: work.index,
			buf:   buf,
		}
	}

	return nil
}

func checkWorkHash(work *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], work.hash[:]) {
		return fmt.Errorf("failed checking integirty for %v", work.index)
	}

	return nil
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download")

	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{
			index:  index,
			hash:   hash,
			length: length,
		}
	}

	// Start the workers
	for _, peer := range t.Peers {
		// Errors are expected when downloading for peers
		// We can ignore them on lint
		go t.startDownloadWorker(peer, workQueue, results) //nolint:errcheck
	}

	// Collect results
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}

	close(workQueue)

	return buf, nil
}

func (t Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLenght
	end = begin + t.PieceLenght
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}
