package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageID uint8

// Types of messages
// More information can me found on https://wiki.theory.org/BitTorrentSpecification#Messages
const (
	MsgChoke         MessageID = 0
	MsgUnchoke       MessageID = 1
	MsgInterrested   MessageID = 2
	MsgNotInterested MessageID = 3
	MsgHave          MessageID = 4
	MsgBitfield      MessageID = 5
	MsgRequest       MessageID = 6
	MsgPiece         MessageID = 7
	MsgCancel        MessageID = 8
)

// Message is the structure of a new peer message
// Formed by the MessageID and a payload
type Message struct {
	ID      MessageID
	Payload []byte
	Buffer  io.Reader
	Length  uint32
}

// NewMessage returns a new message
func NewMessage(ID MessageID, payload []byte) Message {
	return Message{
		ID:      ID,
		Payload: payload,
	}
}

// NewRequestMessage build a new request message
// A request is formed by a index, begin and length
func NewRequestMessage(index, begin, length int) Message {
	// Builds a payload for the message
	payload := make([]byte, 12)
	// index: integer specifying the zero-based piece index
	binary.BigEndian.PutUint32(payload[:4], uint32(index))
	// begin: integer specifying the zero-based byte offset within the piece
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	// length: integer specifying the requested length.
	binary.BigEndian.PutUint32(payload[8:], uint32(length))

	return NewMessage(
		MsgRequest,
		payload,
	)
}

// NewHaveMessage builds a message have
// It accepts a index
func NewHaveMessage(index int) Message {
	payload := make([]byte, 4)
	binary.BigEndian.AppendUint32(payload[:], uint32(index))
	return NewMessage(
		MsgHave,
		payload,
	)
}

// Serialize a message into bytes
// Each message is formed by:
// - Length of the payload as big endian
// - The message ID
// - The payload at the end
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	// Add the length
	length := uint32(len(m.Payload) + 1)
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	// Add the message ID
	buf[4] = byte(m.ID)
	// Finally append the payload
	copy(buf[5:], m.Payload)
	return buf
}

// Read generates a message from a steam
func Unmarshal(r io.Reader) (Message, error) {
	lengthBuf := make([]byte, 4)
	// Reads the first 4 bytes to get the payload length
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return Message{}, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// Keep alive
	if length == 0 {
		return Message{}, nil
	}

	// Read the message ID, a single byte
	messageIDBuf := make([]byte, 1)
	_, err = io.ReadFull(r, messageIDBuf)
	if err != nil {
		return Message{}, err
	}
	messageID := MessageID(messageIDBuf[0])

	// Special handler for Piece messages
	// It may be too big and we don't want to allocate that much memory
	if messageID == MsgPiece {
		return Message{
			ID:     messageID,
			Buffer: r,
			Length: length - 1,
		}, nil
	}

	// Read the message
	messageBuf := make([]byte, length-1) // The length counts the message ID too
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return Message{}, err
	}

	msg := NewMessage(
		messageID,
		messageBuf[:],
	)

	return msg, nil
}

// ParsePiece parses and validates a piece from a index, buffer and a message
// Checks for message type, min payload, index, offset
// returns the
func ParsePiece(index int, buf []byte, msg Message) (int, error) {
	// The id must be Piece
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
	}

	// Read the index from the buffer
	pieceIndex := make([]byte, 4)
	_, err := io.ReadFull(msg.Buffer, pieceIndex)
	if err != nil {
		return 0, err
	}

	// Check if it's the correct index
	parsedIndex := int(binary.BigEndian.Uint32(pieceIndex))
	if parsedIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", index, parsedIndex)
	}

	// Read the index from the buffer
	pieceOffset := make([]byte, 4)
	_, err = io.ReadFull(msg.Buffer, pieceOffset)
	if err != nil {
		return 0, err
	}

	// Check if the the offset is correct
	begin := int(binary.BigEndian.Uint32(pieceOffset))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}

	// Read the data from the buffer, we can mke it go directly to our expected buf
	pieceData := buf[begin : begin+(int(msg.Length)-8)]
	_, err = io.ReadFull(msg.Buffer, pieceData)
	if err != nil {
		return 0, err
	}

	// Check the payload data against the offset and length
	if begin+len(pieceData) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(pieceData), begin, len(buf))
	}

	// Writes the buf to the data at offset
	copy(buf[begin:], pieceData)
	return len(pieceData), nil
}

// ParseHave parses a message have, and returns the index
func ParseHave(msg Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("Expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected payload length 4, got length %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}
