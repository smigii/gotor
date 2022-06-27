package p2p

import (
	"strings"

	"gotor/bf"
)

// MsgBitfieldMinLen is the minimum total length of the message (len 4 + type 1 + payload 0)
const MsgBitfieldMinLen = uint32(5)

// ============================================================================
// TYPES ======================================================================

type MsgBitfield struct {
	msgBase
	bitfield *bf.Bitfield // INCLUDES THE 5 PREFIX BYTES (id<1> + len<4> + bf<X>)
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgBitfield(bf *bf.Bitfield) *MsgBitfield {
	return &MsgBitfield{
		msgBase: msgBase{
			length: uint32(1 + bf.Nbytes()),
			mtype:  TypeBitfield,
		},
		bitfield: bf,
	}
}

// ============================================================================
// GETTER =====================================================================

func (bf *MsgBitfield) Bitfield() *bf.Bitfield {
	return bf.bitfield
}

// ============================================================================
// IMPL =======================================================================

func (bf *MsgBitfield) Encode() []byte {
	data := bf.bitfield.Data5()
	bf.msgBase.fillBase(data[:5])
	return data
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
func DecodeMsgBitfield(fullmsg []byte, msglen uint32) (*MsgBitfield, error) {
	//if uint32(len(fullmsg)) != msglen-1 {
	//	return nil, fmt.Errorf("message length (%v) does not match fullmsg length (%v)", msglen, len(fullmsg))
	//}

	nbytes := int64(4 + msglen)
	nbits := int64(msglen-1) * 8
	bitfield, e := bf.FromBytes(fullmsg[:nbytes], nbits)
	if e != nil {
		return nil, e
	}

	return &MsgBitfield{
		msgBase: msgBase{
			length: msglen,
			mtype:  TypeBitfield,
		},
		bitfield: bitfield,
	}, nil
}
