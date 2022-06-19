/* no_payload.go ==============================================================
Implements the Choke, Unchoke, Interested and NotInterested messages (0 - 3
respectively).
============================================================================ */

package p2p

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

type MsgChoke = msgBase
type MsgUnchoke = msgBase
type MsgInterested = msgBase
type MsgNotInterested = msgBase

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgChoke() *MsgChoke {
	return &MsgChoke{
		length: 1,
		mtype:  TypeChoke,
	}
}

func NewMsgUnchoke() *MsgUnchoke {
	return &MsgUnchoke{
		length: 1,
		mtype:  TypeUnchoke,
	}
}

func NewMsgInterested() *MsgInterested {
	return &MsgInterested{
		length: 1,
		mtype:  TypeInterested,
	}
}

func NewMsgNotInterested() *MsgNotInterested {
	return &MsgNotInterested{
		length: 1,
		mtype:  TypeNotInterested,
	}
}
