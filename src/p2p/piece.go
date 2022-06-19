package p2p

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// MsgPieceMinPayloadLen is the minimum length of a piece message's payload
const MsgPieceMinPayloadLen = uint32(8)

// ============================================================================
// TYPES ======================================================================

type MsgPiece struct {
	msgBase
	index uint32
	begin uint32
	block []byte
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgPiece(index uint32, begin uint32, block []byte) *MsgPiece {
	return &MsgPiece{
		msgBase: msgBase{
			length: uint32(1 + int(MsgPieceMinPayloadLen) + len(block)),
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
	pl := mp.msgBase.Encode()
	binary.BigEndian.PutUint32(pl, mp.index)
	binary.BigEndian.PutUint32(pl, mp.begin)
	pl = append(pl, mp.block...)
	return pl
}

func (mp *MsgPiece) String() string {
	strb := strings.Builder{}
	strb.WriteString(mp.msgBase.String())
	strb.WriteString(fmt.Sprintf("Index: %v\n", mp.index))
	strb.WriteString(fmt.Sprintf("Begin: %v\n", mp.begin))
	strb.WriteString(fmt.Sprintf("Block: %v\n", mp.block))
	return strb.String()
}

// ============================================================================
// FUNC =======================================================================

func DecodeMsgPiece(payload []byte, length uint32) (*MsgPiece, error) {
	if uint32(len(payload)) != length {
		return nil, fmt.Errorf("piece message must have %v byte payload, got %v", MsgHavePayloadLen, len(payload))
	}
	if uint32(len(payload)) < MsgPieceMinPayloadLen {
		return nil, fmt.Errorf("piece message payload must be at least %v bytes, got %v", MsgPieceMinPayloadLen, len(payload))
	}
	index := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	block := payload[8:]
	return NewMsgPiece(index, begin, block), nil
}
