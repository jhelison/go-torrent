package client

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"runtime"

	"github.com/jhelison/go-torrent/filesystem"
	"github.com/jhelison/go-torrent/logger"
	"github.com/jhelison/go-torrent/marshallers/handshake"
	"github.com/jhelison/go-torrent/marshallers/peer"
)

var (
	// Default logger
	log = logger.GetLogger()
)

// Torrent is the full representation for a torrent with peers and pieces
type Torrent struct {
	Peers       []peer.Peer
	PeerID      handshake.PeerID
	InfoHash    handshake.Hash
	PieceHashes []handshake.Hash
	PieceLength int
	Length      int
	Name        string
}

// pieceWork is a single work from a piece
type pieceWork struct {
	index  int
	hash   handshake.Hash
	length int
}

// pieceResult if the final result from a piece
type pieceResult struct {
	index int
	buf   []byte
}

// startDownloadWorker start a new worker to download a piece from a peer
func (t *Torrent) startDownloadWorker(peer peer.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	// Create a new client for the peer
	client, err := NewClient(peer, t.PeerID, t.InfoHash)
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

	for work := range workQueue {
		// Check if the client has been banned before new work
		if client.banned {
			log.Error().Msgf("peer %s has been banned", peer)
			workQueue <- work
			return
		}

		// Create a new state for that piece work
		state := pieceState{
			work:   work,
			client: client,
			buf:    make([]byte, work.length),
		}

		err := state.processPiece()
		if err != nil {
			log.Warn().Msgf("Error when processing piece %v, err: %s", work.index, err)
			// If any error happens we can try the work again
			state = pieceState{}
			workQueue <- work
			continue
		}

		// Send that now we have that piece
		err = client.SendHave(work.index)
		if err != nil {
			log.Warn().Msgf("sending has failed, err: %s", err)
		}

		// Append the downloaded piece to the results
		results <- &pieceResult{
			index: work.index,
			buf:   state.buf,
		}
		state = pieceState{}
	}
}

// checkWorkHash takes a single buf and check it's sha1 hash against
// expected work hash
func checkWorkHash(work *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], work.hash[:]) {
		return fmt.Errorf("failed checking integrity for %v", work.index)
	}

	return nil
}

// Download downloads a torrent
func (t *Torrent) Download(path string) error {
	log.Info().Msg("Starting download")
	log.Info().Msgf("Total available peers: %v", len(t.Peers))

	filePath := fmt.Sprintf("%s/%s", path, t.Name)

	// Create a new work queue and result that are shared between peers
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

	// Create a new file
	file, err := filesystem.CreateFileWithSize(filePath, int64(t.Length))
	if err != nil {
		return err
	}
	defer file.Close()

	// Start the workers, one per each peer
	for _, peer := range t.Peers {
		// Errors are expected when downloading for peers
		// We can ignore them on lint
		go t.startDownloadWorker(peer, workQueue, results)
	}

	// Collect results
	donePieces := 0
	// Keep iterating until we are done with the pieces
	for donePieces < len(t.PieceHashes) {
		// Take the result, calculate the boundaries and safe on the buf
		res := <-results
		begin, _ := t.calculateBoundsForPiece(res.index)
		err := filesystem.WriteFileChunk(file, res.buf, int64(begin))
		if err != nil {
			return err
		}
		donePieces++

		// Log to user
		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		missingPieces := len(t.PieceHashes) - donePieces
		log.Info().Msgf("(%0.2f%%) Downloaded piece #%d from %d peers, missing %v from %v pieces", percent, res.index, numWorkers, missingPieces, len(t.PieceHashes))
	}

	close(workQueue)

	// Return the final buffer
	return nil
}

// calculatedPieceSize calculated a piece size for a index
func (t Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

// calculateBoundsForPiece calculates the boundaries for a piece
// returns a begin and a end
// The begin is the pieceLength multiplied by the index
// The end is the begin + the pieceLength with the end as a threshold
func (t Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}
