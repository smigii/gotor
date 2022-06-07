package protocol

import (
	"encoding/binary"
	"fmt"
)

// LenHave is the length of the message as defined in BEP_0003
const LenHave = uint32(5)

// LenHavePayload is the payload size in bytes
const LenHavePayload = uint32(4)

// ============================================================================
// TYPES ======================================================================

type MsgHave struct {
	MsgBase
	index uint32
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgHave(idx uint32) *MsgHave {
	return &MsgHave{
		MsgBase: MsgBase{
			length: LenHave,
			mtype:  TypeHave,
		},
		index: idx,
	}
}

// ============================================================================
// GETTER =====================================================================

func (h *MsgHave) PieceIdx() uint32 {
	return h.index
}

// ============================================================================
// IMPL =======================================================================

func (h *MsgHave) Encode() []byte {
	pl := h.MsgBase.Encode()
	binary.BigEndian.PutUint32(pl, h.index)
	return pl
}

// ============================================================================
// FUNC =======================================================================

func DecodeMsgHave(payload []byte) (*MsgHave, error) {
	if uint32(len(payload)) != LenHavePayload {
		return nil, fmt.Errorf("have message must have %v byte payload, got %v", LenHavePayload, len(payload))
	}
	idx := binary.BigEndian.Uint32(payload)
	return NewMsgHave(idx), nil
}
