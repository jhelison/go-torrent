package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"go-torrent/client"
	"go-torrent/logger"
	"go-torrent/message"
	"go-torrent/peers"
	"runtime"
	"time"
)

var (
	// Client configurations
	DownloadDeadline = 5 * time.Second
	MaxBacklog       = 5
	MaxBlockSize     = 16384

	// Default logger
	log = logger.GetLogger()
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

	// Set a deadline to skip stuck peers
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

// startDownloadWorker start a new worker to download a piece from a peer
func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	// Create a new client for the peer
	client, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Warn().Msgf("failed to start handshake with peer %s, err: %s", peer, err)
		return
	}
	defer client.Conn.Close()

	log.Info().Msgf("Handshake complete with peer %s", peer)

	// Send unchoke
	err = client.SendUnchoke()
	if err != nil {
		log.Warn().Msgf("failed to send unchoke to peer %s, err: %s", peer, err)
		return
	}

	// Send that the client is interested
	err = client.SendInterested()
	if err != nil {
		log.Warn().Msgf("failed to send interested to peer %s, err: %s", peer, err)
		return
	}

	// TODO: Turn this into a env var
	// Count of retries for the peer
	retries := 0
	for work := range workQueue {
		// Check if we have reached max retries
		if retries >= 30 {
			workQueue <- work
			log.Warn().Msgf("max retries reached for peer %s", peer)
			return
		}

		// Check if worker has piece
		if !client.Bitfield.HasPiece(work.index) {
			log.Trace().Msgf("piece %s not found on peer %s", work.hash, peer)
			workQueue <- work
			continue
		}

		// Download piece from peers
		buf, err := downloadPiece(client, work)
		if err != nil {
			log.Warn().Msgf("error downloading piece with peer %s, err: %s", peer, err)
			workQueue <- work
			retries++
			continue
		}

		// Check the integrity
		err = checkWorkHash(work, buf)
		if err != nil {
			log.Warn().Msgf("integrity validation failed, invalid index: %v", work.index)
		}

		// Send that now we have that piece
		err = client.SendHave(work.index)
		if err != nil {
			log.Warn().Msgf("sending has failed, err: %s", err)
		}

		// Append the downloaded piece to the results
		results <- &pieceResult{
			index: work.index,
			buf:   buf,
		}
	}
}

func checkWorkHash(work *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], work.hash[:]) {
		return fmt.Errorf("failed checking integirty for %v", work.index)
	}

	return nil
}

func (t *Torrent) Download() ([]byte, error) {
	log.Info().Msg("Starting download")
	log.Info().Msgf("Total available peers: %v", len(t.Peers))

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
		missingPieces := len(t.PieceHashes) - donePieces
		log.Info().Msgf("(%0.2f%%) Downloaded piece #%d from %d peers, missing %v from %v pieces", percent, res.index, numWorkers, missingPieces, len(t.PieceHashes))
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
