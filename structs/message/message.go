package message

import (
	"encoding/binary"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

// the output message has the form
// <length prefix><message ID><payload>
// length == 4 bytes
// message ID == 1 byte
// paylod == length - 1
func (m *Message) Serialize() []byte {
	if m == nil {
		return nil
	}
	length := uint32(len(m.Payload) + 1)
	buf := make([]byte, 4+length)
	cur := 0
	binary.BigEndian.PutUint32(buf[cur:], length)
	cur += 4
	buf[cur] = byte(m.ID)
	copy(buf[cur+1:], m.Payload)
	return buf
}

// Read reads a message from the reader
// return Message object and error if there is a error
// return nil and nil if there is no message
func Read(r io.Reader) (*Message, error) {
	length := make([]byte, 4)
	_, err := io.ReadFull(r, length)
	if err != nil {
		return nil, err
	}
	lengthInt := binary.BigEndian.Uint32(length)
	if lengthInt == 0 {
		return nil, nil
	}

	msg := make([]byte, lengthInt)
	_, err = io.ReadFull(r, msg)
	lengthInt -= 1
	if err != nil {
		return nil, err
	}
	payload := make([]byte, lengthInt)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:      messageID(msg[0]),
		Payload: payload,
	}, nil
}
