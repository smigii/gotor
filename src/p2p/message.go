package p2p

import (
	"encoding/binary"
	"fmt"
	"strings"
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
	TypeKeepAlive = uint8(255) // Not in BEP_0003, however makes checking easy
)

// MsgLengthPrefixLen is the length of the 4 byte <length prefix> of all p2p messages
const MsgLengthPrefixLen = uint8(4)

// PayloadStart is the starting index of the payload for all p2p messages
const PayloadStart = uint32(5)

// ============================================================================
// TYPES ======================================================================

type Message interface {
	Length() uint32 // Length of the message's ID and payload (1 + X). Excludes 4 byte length
	Mtype() uint8
	Encode() []byte
	String() string
}

// DecodeAllResult holds the decoded messages, as well as the number of bytes
// that were succesfully read out of the given byte slice.
type DecodeAllResult struct {
	Msgs []Message
	Read uint64 // Number of bytes succesfully read
}

// DecodeResult holds the decoded Message and the number of bytes that were
// read from the given byte slice.
type DecodeResult struct {
	Msg  Message
	Read uint64 // Number of bytes read
}

type msgBase struct {
	length uint32
	mtype  uint8
}

// ============================================================================
// IMPL =======================================================================

func (m *msgBase) Length() uint32 {
	return m.length
}

func (m *msgBase) Mtype() uint8 {
	return m.mtype
}

func (m *msgBase) Encode() []byte {
	pl := make([]byte, 5, 5)
	binary.BigEndian.PutUint32(pl, m.length)
	pl[4] = m.mtype

	return pl
}

func (m *msgBase) String() string {
	strb := strings.Builder{}
	strb.WriteString("Base Message\n")
	strb.WriteString(fmt.Sprintf("  Type: %v\n", m.mtype))
	strb.WriteString(fmt.Sprintf("Length: %v", m.length))
	return strb.String()
}

// ============================================================================
// FUNC =======================================================================

// fillBase should be used by structs that implement the Message interface to
// fill the output byte slice (buf) when encoding. Use this instead of calling
// msgBase.Encode() to avoid allocating multiple byte slices.
func (m *msgBase) fillBase(buf []byte) {
	_ = buf[4] // bounds check hint to compiler; see golang.org/issue/14808
	binary.BigEndian.PutUint32(buf, m.length)
	buf[4] = m.mtype
}

// DecodeAll reads through all the messages encoded in the data byte slice and
// returns a DecodeAllResult containing the decoded Message struct and the
// number of bytes read from the data byte slice.
func DecodeAll(data []byte) DecodeAllResult {
	msgs := make([]Message, 0, 8)

	idx := uint64(0)
	read := uint64(0)

	for {
		if idx >= uint64(len(data)) {
			break
		}
		dr, err := Decode(data[idx:])
		if err == nil {
			msgs = append(msgs, dr.Msg)
			read += dr.Read
		}
		idx += dr.Read
	}

	return DecodeAllResult{
		Msgs: msgs,
		Read: read,
	}
}

// Decode decodes a single Message from the provided byte slice data. Returns a
// DecodeResult containing the decoded Message and the number of bytes read
// from the byte slice.
func Decode(data []byte) (DecodeResult, error) {

	// Error value
	// For now, force caller to discard remaining data. If this function
	// throws an error, the data should be assumed to be garbage. Setting
	// Read to len(data) means the caller won't know how much data is
	// left.
	badResult := DecodeResult{
		Msg:  nil,
		Read: uint64(len(data)),
	}

	// All messages have a 4 byte length prefix
	datalen := len(data)
	if datalen < int(MsgLengthPrefixLen) {
		return badResult, fmt.Errorf("message length must be at least 4, got %v", datalen)
	}

	// Get message length
	msglen := binary.BigEndian.Uint32(data[:MsgLengthPrefixLen])

	// Len 0000 indicates keep alive message
	if msglen == 0 {
		return DecodeResult{
			Msg:  &KeepAliveSingleton,
			Read: uint64(MsgKeepAliveTotalLen),
		}, nil
	}

	// Get message type
	if datalen < 5 {
		return badResult, fmt.Errorf("non-keep-alive message length must be at least 5, got %v", datalen)
	}
	mtype := uint8(data[4])

	if mtype <= 3 && msglen != 1 {
		// Check invalid message for no-payload messages
		// Type 0 = Choke
		// Type 1 = Unchoke
		// Type 2 = Interested
		// Type 3 = Not Interested
		return badResult, fmt.Errorf("invalid message, length for id [0-3] must be 1, got %v", msglen)
	} else if uint32(len(data)) < 4+msglen {
		// Messages with payload
		return badResult, fmt.Errorf("length specified as %v, payload length is %v", msglen, len(data))
	}

	// msglen includes length byte, -1 to exclude it
	payload := data[PayloadStart : PayloadStart+msglen-1]

	var msg Message
	var n uint64
	var err error

	// Messages with no payload
	switch mtype {
	case TypeChoke:
		msg = NewMsgChoke()
		n = uint64(MsgChokeTotalLen)
		err = nil
	case TypeUnchoke:
		msg = NewMsgUnchoke()
		n = uint64(MsgUnchokeTotalLen)
		err = nil
	case TypeInterested:
		msg = NewMsgInterested()
		n = uint64(MsgInterestedTotalLen)
		err = nil
	case TypeNotInterested:
		msg = NewMsgNotInterested()
		n = uint64(MsgNotInterestedTotalLen)
		err = nil
	case TypeHave:
		msg, err = DecodeMsgHave(payload)
		n = uint64(MsgHaveTotalLen)
	case TypeBitfield:
		msg, err = DecodeMsgBitfield(data, msglen)
		if err != nil {
			n = uint64(uint32(MsgLengthPrefixLen) + msg.Length())
		}
	case TypeRequest:
		msg, err = DecodeMsgRequest(payload)
		n = uint64(MsgRequestTotalLen)
	case TypePiece:
		msg, err = DecodeMsgPiece(payload, msglen)
		n = uint64(uint32(MsgLengthPrefixLen) + msg.Length())
	default:
		msg = nil
		err = fmt.Errorf("unknown message type %v", mtype)
		// Force DecodeResult to be unusable, entire data byte slice should be discarded
		n = uint64(len(data))
	}

	return DecodeResult{
		Msg:  msg,
		Read: n,
	}, err
}
