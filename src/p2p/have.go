/* have.go ====================================================================
Implements the have protocol message
============================================================================ */

package p2p

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	// MsgHaveTotalLen is the total length of the message in bytes (len 4 + type 1 + payload 4)
	MsgHaveTotalLen = uint32(9)

	// MsgHaveSpecLen is the length of the message as defined in BEP_0003 (id 1 + payload 4)
	MsgHaveSpecLen = uint32(5)

	// MsgHavePayloadLen is the payload size in bytes
	MsgHavePayloadLen = uint32(4)
)

// ============================================================================
// TYPES ======================================================================

type MsgHave struct {
	msgBase
	index uint32
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgHave(idx uint32) *MsgHave {
	return &MsgHave{
		msgBase: msgBase{
			length: MsgHaveSpecLen,
			mtype:  TypeHave,
		},
		index: idx,
	}
}

// ============================================================================
// GETTER =====================================================================

func (h *MsgHave) Index() uint32 {
	return h.index
}

// ============================================================================
// IMPL =======================================================================

func (h *MsgHave) Encode() []byte {
	pl := make([]byte, MsgHaveTotalLen, MsgHaveTotalLen)
	h.msgBase.fillBase(pl)
	binary.BigEndian.PutUint32(pl[PayloadStart:MsgHaveTotalLen], h.index)
	return pl
}

func (h *MsgHave) String() string {
	strb := strings.Builder{}
	strb.WriteString(h.msgBase.String())
	strb.WriteString(fmt.Sprintf("Index: %v\n", h.index))
	return strb.String()
}

// ============================================================================
// FUNC =======================================================================

func DecodeMsgHave(payload []byte) (*MsgHave, error) {
	if uint32(len(payload)) != MsgHavePayloadLen {
		return nil, fmt.Errorf("have message must have %v byte payload, got %v", MsgHavePayloadLen, len(payload))
	}
	idx := binary.BigEndian.Uint32(payload)
	return NewMsgHave(idx), nil
}
