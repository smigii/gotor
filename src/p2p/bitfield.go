package p2p

// ============================================================================
// TYPES ======================================================================

type MsgBitfield struct {
	MsgBase
	bitfield []byte
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgBitfield(bitfield []byte) *MsgBitfield {
	return &MsgBitfield{
		MsgBase: MsgBase{
			length: uint32(len(bitfield) + 1),
			mtype:  TypeHave,
		},
		bitfield: bitfield,
	}
}

// ============================================================================
// IMPL =======================================================================

func (h *MsgBitfield) Encode() []byte {
	pl := h.MsgBase.Encode()
	pl = append(pl, h.bitfield...)
	return pl
}
