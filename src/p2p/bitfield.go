package p2p

import (
	"fmt"
	"strings"
)

// MsgBitfieldMinLen is the minimum total length of the message (len 4 + type 1 + payload 0)
const MsgBitfieldMinLen = uint32(5)

// ============================================================================
// TYPES ======================================================================

type MsgBitfield struct {
	msgBase
	bitfield []byte // INCLUDES THE 5 PREFIX BYTES (id<1> + len<4> + bf<X>)
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgBitfield(bitfield []byte) *MsgBitfield {
	return &MsgBitfield{
		msgBase: msgBase{
			length: 1 + uint32(len(bitfield)),
			mtype:  TypeBitfield,
		},
		bitfield: bitfield,
	}
}

// ============================================================================
// GETTER =====================================================================

func (bf *MsgBitfield) Bitfield() []byte {
	return bf.bitfield
}

// ============================================================================
// IMPL =======================================================================

func (bf *MsgBitfield) Encode() []byte {
	// TODO: This may need optimization.
	// 4GiB torrents seem to have around 1k pieces, meaning we are
	// copying over 1000 elements every time we call Encode(). For
	// larger torrents, this could cause some bad performance.
	bflen := uint32(len(bf.bitfield))
	pl := make([]byte, MsgBitfieldMinLen, MsgBitfieldMinLen+bflen)
	bf.msgBase.fillBase(pl)
	pl = append(pl, bf.bitfield...)
	return pl
}

func (bf *MsgBitfield) String() string {
	strb := strings.Builder{}
	strb.WriteString("Message: Bitfield\n")
	return strb.String()
}

// ============================================================================
// FUNC =======================================================================

// DecodeMsgBitfield does what you think. It explicitly asks for the length
// of the bitfield, you should pass the length that was encoded in the
// full message, to ensure that the entire bitfield is being stored.
func DecodeMsgBitfield(bitfield []byte, msglen uint32) (*MsgBitfield, error) {
	if uint32(len(bitfield)) != msglen-1 {
		return nil, fmt.Errorf("message length (%v) does not match bitfield length (%v)", msglen, len(bitfield))
	}
	return &MsgBitfield{
		msgBase: msgBase{
			length: msglen,
			mtype:  TypeBitfield,
		},
		bitfield: bitfield,
	}, nil
}