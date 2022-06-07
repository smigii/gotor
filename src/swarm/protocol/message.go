package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	TypeChoke = uint8(iota)
	TypeUnchoke
	TypeInterested
	TypeNotInterested
	TypeHave
	TypeBitfield
	TypeRequest
	TypePiece
	TypeCancel
)

// MaxMsgType is the max value for a message type given by spec (BEP_0003).
const MaxMsgType = uint8(8)

// KeepAlive is a global keep alive message. We will be sending lots of these,
// doesn't make sense to allocate a new one every time.
var KeepAlive = MsgBase{
	length: 0,
	mtype:  0,
}

// ============================================================================
// TYPES ======================================================================

type Message interface {
	Length() uint32
	Mtype() uint8
	Encode() []byte
}

type MsgBase struct {
	length uint32
	mtype  uint8
}

// ============================================================================
// IMPL =======================================================================

func (m *MsgBase) Length() uint32 {
	return m.length
}

func (m *MsgBase) Mtype() uint8 {
	return m.mtype
}

func (m *MsgBase) Encode() []byte {
	pl := make([]byte, 0, 5)
	binary.BigEndian.PutUint32(pl, m.length)
	pl[4] = m.mtype

	return pl
}

// ============================================================================
// FUNC =======================================================================

// DecodeAll reads through all the messages encoded in the data byte slice
// and returns all the messages and errors it encountered when reading.
func DecodeAll(data []byte) ([]*MsgBase, []error) {

	return nil, nil
}

// Decode returns a single message or error from the data byte slice
func Decode(data []byte) (Message, error) {

	// All messages have a 4 byte length prefix
	datalen := len(data)
	if datalen < 4 {
		return nil, fmt.Errorf("message length must be at least 4, got %v", datalen)
	}

	length := binary.BigEndian.Uint32(data[:4])

	// Len 0000 indicates keep alive message
	if length == 0 {
		return &KeepAlive, nil
	}

	if datalen < 5 {
		return nil, fmt.Errorf("non-keep-alive message length must be at least 5, got %v", datalen)
	}
	mtype := uint8(data[4])

	// Messages with no payload
	switch mtype {
	case TypeChoke:
		return NewMsgChoke(), nil
	case TypeUnchoke:
		return NewMsgUnchoke(), nil
	case TypeInterested:
		return NewMsgInterested(), nil
	case TypeNotInterested:
		return NewMsgNotInterested(), nil
	}

	// Messages with payload
	if uint32(len(data)) < 4+length {
		return nil, fmt.Errorf("length specified as %v, payload length is %v", length, len(data))
	}

	// Exclude the length (byte 4)
	payload := data[5 : length-1]

	switch mtype {
	case TypeHave:
		msg, err := DecodeMsgHave(payload)
		return msg, err
	case TypeBitfield:
		return NewMsgBitfield(payload), nil
	case TypeRequest:
		msg, err := DecodeMsgRequest(payload)
		return msg, err
	case TypePiece:
		msg, err := DecodeMsgPiece(payload, length)
		return msg, err
	default:
		return nil, fmt.Errorf("unknown message type %v", mtype)
	}
}
