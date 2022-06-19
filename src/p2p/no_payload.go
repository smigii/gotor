/* no_payload.go ==============================================================
Implements the Choke, Unchoke, Interested and NotInterested messages (0 - 3
respectively).
============================================================================ */

package p2p

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
