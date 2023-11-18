package message

import (
	"encoding/binary"
	"io"
)

type MessageID uint8

const (
	MsgChoke          MessageID = 0
	MsgUnchoke        MessageID = 1
	MsgInterrested    MessageID = 2
	MsgNotInterrested MessageID = 3
	MsgHave           MessageID = 4
	MsgBitfield       MessageID = 5
	MsgRequest        MessageID = 6
	MsgPiece          MessageID = 7
	MsgCancel         MessageID = 8
)

type Message struct {
	ID      MessageID
	Payload []byte
}

func NewMessage(ID MessageID, payload []byte) Message {
	return Message{
		ID:      ID,
		Payload: payload,
	}
}

func NewRequestMessage(index, begin, legth int) Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:], uint32(legth))
	return NewMessage(
		MsgRequest,
		payload,
	)
}

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

	// Add the lenght
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
func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	// Reads the first 4 bytes to get the payload lenght
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// Keep alive
	if length == 0 {
		return nil, nil
	}

	// Read the message
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	msg := NewMessage(
		MessageID(messageBuf[0]),
		messageBuf[1:],
	)

	return &msg, nil
}
