package protocol

import (
	"encoding/binary"
	"fmt"
)

// MinLenPiece is the minimum length of a piece message's payload
const MinLenPiece = uint32(8)

// ============================================================================
// TYPES ======================================================================

type MsgPiece struct {
	MsgBase
	index uint32
	begin uint32
	block []byte
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgPiece(index uint32, begin uint32, block []byte) *MsgPiece {
	return &MsgPiece{
		MsgBase: MsgBase{
			length: LenHave,
			mtype:  TypeHave,
		},
		index: index,
		begin: begin,
		block: block,
	}
}

// ============================================================================
// GETTER =====================================================================

func (mp *MsgPiece) Index() uint32 {
	return mp.index
}

func (mp *MsgPiece) Begin() uint32 {
	return mp.begin
}

func (mp *MsgPiece) Block() []byte {
	return mp.block
}

// ============================================================================
// IMPL =======================================================================

func (mp *MsgPiece) Encode() []byte {
	pl := mp.MsgBase.Encode()
	binary.BigEndian.PutUint32(pl, mp.index)
	binary.BigEndian.PutUint32(pl, mp.begin)
	pl = append(pl, mp.block...)
	return pl
}

// ============================================================================
// FUNC =======================================================================

func DecodeMsgPiece(payload []byte, length uint32) (*MsgPiece, error) {
	if uint32(len(payload)) != length {
		return nil, fmt.Errorf("piece message must have %v byte payload, got %v", LenHavePayload, len(payload))
	}
	if uint32(len(payload)) < MinLenPiece {
		return nil, fmt.Errorf("piece message payload must be at least %v bytes, got %v", MinLenPiece, len(payload))
	}
	index := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	block := payload[8:]
	return NewMsgPiece(index, begin, block), nil
}
