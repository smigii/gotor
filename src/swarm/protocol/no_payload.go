package protocol

// ============================================================================
// TYPES ======================================================================

type MsgChoke = MsgBase
type MsgUnchoke = MsgBase
type MsgInterested = MsgBase
type MsgNotInterested = MsgBase

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
