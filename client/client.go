package client

import (
	"bytes"
	"fmt"
	"go-torrent/bitfield"
	"go-torrent/message"
	"go-torrent/peers"
	"net"
	"time"
)

var (
	TimoutTime    = 3 * time.Second
	TimoutTimeBig = 5 * time.Second
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash peers.Hash
	peerID   peers.PeerID
}

func New(peer peers.Peer, peerID peers.PeerID, infoHash peers.Hash) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), TimoutTime)
	if err != nil {
		return nil, err
	}

	_, err = completeHandshake(conn, peerID, infoHash)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := recieveBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}

func completeHandshake(conn net.Conn, peerID peers.PeerID, infoHash peers.Hash) (*peers.Handshake, error) {
	err := conn.SetDeadline(time.Now().Add(TimoutTime))
	if err != nil {
		return nil, err
	}
	// Disable the deadline at the end
	// We can ignore the error
	defer conn.SetDeadline(time.Time{}) //nolint:errcheck

	// Do the handshake
	handshake := peers.NewHandshake(peerID, infoHash)
	_, err = conn.Write(handshake.Serialize())
	if err != nil {
		return nil, err
	}

	// Read the response handshake
	res, err := peers.ReadHandshake(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(infoHash[:], res.InfoHash[:]) {
		return nil, fmt.Errorf("expected info hash %s but got %s", infoHash, res.InfoHash)
	}

	return res, nil
}

func recieveBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	err := conn.SetDeadline(time.Now().Add(TimoutTimeBig))
	if err != nil {
		return nil, err
	}
	// We can ignore the error for this line
	defer conn.SetDeadline(time.Time{}) //nolint:errcheck

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, fmt.Errorf("expected bitfield but got %v", msg)
	}
	if msg.ID != message.MsgBitfield {
		return nil, fmt.Errorf("expected bitfield but got %v", msg)
	}

	return msg.Payload, nil
}

func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

func (c *Client) SendRequest(index, begin, length int) error {
	msg := message.NewRequestMessage(index, begin, length)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(index int) error {
	msg := message.NewHaveMessage(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := message.NewMessage(message.MsgInterrested, nil)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	msg := message.NewMessage(message.MsgNotInterrested, nil)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := message.NewMessage(message.MsgUnchoke, nil)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
