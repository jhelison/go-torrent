package client

import (
	"fmt"
	"time"

	"go-torrent/marshallers/message"
)

// pieceState is the representation of the current progress of a single piece
type pieceState struct {
	client     *Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
	work       *pieceWork
}

// processPiece process a single piece and download it
func (state *pieceState) processPiece() error {
	// Check if we have reached max retries
	if state.client.retries >= 5 {
		state.client.banned = true
		return fmt.Errorf("max retries reached for peer %s", state.client.peer)
	}

	// Check if worker has piece
	if !state.client.Bitfield.HasPiece(state.work.index) {
		state.client.retries++
		return fmt.Errorf("piece %s not found on peer %s", state.work.hash, state.client.peer)
	}

	// Download piece from peers
	err := state.downloadPiece()
	if err != nil {
		state.client.retries++
		return fmt.Errorf("error downloading piece with peer %s, err: %s", state.client.peer, err)
	}

	// Check the hash for the buf
	err = checkWorkHash(state.work, state.buf)
	if err != nil {
		state.client.retries++
		return fmt.Errorf("integrity validation failed, invalid index: %v", state.work.index)
	}

	return nil
}

// downloadPiece download a piece from a peer
func (state *pieceState) downloadPiece() error {

	// Set a deadline to skip stuck peers
	err := state.client.Conn.SetDeadline(time.Now().Add(DownloadDeadline))
	if err != nil {
		return err
	}
	// We can ignore the error on the defer
	defer state.client.Conn.SetDeadline(time.Time{}) //nolint:errcheck

	for state.downloaded < state.work.length {
		// If choked we wait a bit
		if state.client.Choked {
			log.Trace().Msgf("Peer %d chocked, waiting a bit", state.client.peer)
			time.Sleep(time.Second)
		} else {
			// We can open request messages until we reach the max backlog
			for state.backlog < MaxBacklog && state.requested < state.work.length {
				blockSize := MaxBlockSize
				// Last block may be shorter
				if state.work.length-state.requested < blockSize {
					blockSize = state.work.length - state.requested
				}

				// Request and create a new backlog
				err := state.client.SendRequest(state.work.index, state.requested, blockSize)
				if err != nil {
					return err
				}

				state.backlog++
				state.requested += blockSize
			}
		}

		// Read a message
		// This can unchoke the client
		err := state.readMessage()
		if err != nil {
			return err
		}
	}

	return nil
}

// readMessage reads a message and update the pieceProgress state
func (state *pieceState) readMessage() error {
	// Read a message from the client
	msg, err := state.client.Read()
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	// Update the state based on the message id
	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		// If the message is have we parse it and update the bitfield with the index
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		// If we have a piece message we parse if and update the state with
		// a new download complete and a smaller backlog
		n, err := message.ParsePiece(state.work.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}
