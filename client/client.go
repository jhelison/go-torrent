package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"go-torrent/marshallers/handshake"
	"go-torrent/marshallers/message"
	"go-torrent/marshallers/peer"

	"github.com/spf13/viper"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield Bitfield
	peer     peer.Peer
	banned   bool
	retries  int
	infoHash handshake.Hash
	peerID   handshake.PeerID
}

// NewClient returns a new client
// This also executes the handshake
func NewClient(
	peer peer.Peer,
	peerID handshake.PeerID,
	infoHash handshake.Hash,
) (*Client, error) {
	// Viper config
	timeout := viper.GetDuration("peers.timeout")

	// Do the tcp dial
	conn, err := net.DialTimeout("tcp", peer.String(), timeout)
	if err != nil {
		return nil, err
	}

	// Complete the handshake with the peer
	_, err = completeHandshake(conn, peerID, infoHash)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Receives the bitfield
	bf, err := recieveBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		banned:   false,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}

// completeHandshake does a handshake with a peer
func completeHandshake(conn net.Conn, peerID handshake.PeerID, infoHash handshake.Hash) (*handshake.Handshake, error) {
	// Viper config
	timeout := viper.GetDuration("peers.timeout")

	err := conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}
	// Disable the deadline at the end
	// We can ignore the error
	defer conn.SetDeadline(time.Time{}) //nolint:errcheck

	// Do the handshakeRes
	handshakeRes := handshake.NewHandshake(peerID, infoHash)
	_, err = conn.Write(handshakeRes.Marshal())
	if err != nil {
		return nil, err
	}

	// Read the response handshake
	res, err := handshake.Unmarshal(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(infoHash[:], res.InfoHash[:]) {
		return nil, fmt.Errorf("expected info hash %s but got %s", infoHash, res.InfoHash)
	}

	return res, nil
}

// recieveBitfield receives a bitfield from a peer
func recieveBitfield(conn net.Conn) (Bitfield, error) {
	// Viper config
	timeout := viper.GetDuration("peers.timeout")

	err := conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}
	// We can ignore the error for this line
	defer conn.SetDeadline(time.Time{}) //nolint:errcheck

	// Validates if we have received a response and it's a bitfield msg
	msg, err := message.Unmarshal(conn)
	if err != nil {
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		return nil, fmt.Errorf("expected bitfield but got %v", msg)
	}

	return msg.Payload, nil
}

// Read reads the message from the client
func (c *Client) Read() (message.Message, error) {
	msg, err := message.Unmarshal(c.Conn)
	return msg, err
}

// SendRequest sends a new request with the expected index, begin and length
func (c *Client) SendRequest(index, begin, length int) error {
	msg := message.NewRequestMessage(index, begin, length)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave send a new have message with a index
func (c *Client) SendHave(index int) error {
	msg := message.NewHaveMessage(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendInterested send a interested message
func (c *Client) SendInterested() error {
	msg := message.NewMessage(message.MsgInterrested, nil)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested sends a not interested messaged
func (c *Client) SendNotInterested() error {
	msg := message.NewMessage(message.MsgNotInterested, nil)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke send a new unchoke message
func (c *Client) SendUnchoke() error {
	msg := message.NewMessage(message.MsgUnchoke, nil)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
