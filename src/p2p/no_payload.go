/* no_payload.go ==============================================================
Implements the Choke, Unchoke, Interested and NotInterested messages (0 - 3
respectively).
============================================================================ */

package p2p

import "strings"

const (
	// MsgChokeTotalLen is the total length of choke (4 len + 1 type)
	MsgChokeTotalLen = uint8(5)

	// MsgUnchokeTotalLen is the total length of choke (4 len + 1 type)
	MsgUnchokeTotalLen = uint8(5)

	// MsgInterestedTotalLen is the total length of choke (4 len + 1 type)
	MsgInterestedTotalLen = uint8(5)

	// MsgNotInterestedTotalLen is the total length of choke (4 len + 1 type)
	MsgNotInterestedTotalLen = uint8(5)
)

// ============================================================================
// TYPES ======================================================================

type MsgChoke struct{ msgBase }
type MsgUnchoke struct{ msgBase }
type MsgInterested struct{ msgBase }
type MsgNotInterested struct{ msgBase }

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgChoke() *MsgChoke {
	return &MsgChoke{
		msgBase: msgBase{
			length: 1,
			mtype:  TypeChoke,
		},
	}
}

func NewMsgUnchoke() *MsgUnchoke {
	return &MsgUnchoke{
		msgBase: msgBase{
			length: 1,
			mtype:  TypeUnchoke,
		},
	}
}

func NewMsgInterested() *MsgInterested {
	return &MsgInterested{
		msgBase: msgBase{
			length: 1,
			mtype:  TypeInterested,
		},
	}
}

func NewMsgNotInterested() *MsgNotInterested {
	return &MsgNotInterested{
		msgBase: msgBase{
			length: 1,
			mtype:  TypeNotInterested,
		},
	}
}

func (m *MsgChoke) String() string {
	strb := strings.Builder{}
	strb.WriteString("Message: Choke")
	return strb.String()
}

func (m *MsgUnchoke) String() string {
	strb := strings.Builder{}
	strb.WriteString("Message: Unchoke")
	return strb.String()
}

func (m *MsgInterested) String() string {
	strb := strings.Builder{}
	strb.WriteString("Message: Interested")
	return strb.String()
}

func (m *MsgNotInterested) String() string {
	strb := strings.Builder{}
	strb.WriteString("Message: Not Interested")
	return strb.String()
}
