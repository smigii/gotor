package p2p

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// LenRequest is the length of the message as defined in BEP_0003
const LenRequest = uint32(13)

// LenRequestPayload is the payload size in bytes
const LenRequestPayload = uint32(12)

// ============================================================================
// TYPES ======================================================================

type MsgRequest struct {
	msgBase
	index  uint32
	begin  uint32
	reqlen uint32 // Length of request, not message
}

// ============================================================================
// CONSTRUCTORS ===============================================================

func NewMsgRequest(index uint32, begin uint32, reqlen uint32) *MsgRequest {
	return &MsgRequest{
		msgBase: msgBase{
			length: LenRequest,
			mtype:  TypeRequest,
		},
		index:  index,
		begin:  begin,
		reqlen: reqlen,
	}
}

// ============================================================================
// GETTER =====================================================================

func (h *MsgRequest) Index() uint32 {
	return h.index
}

func (h *MsgRequest) Begin() uint32 {
	return h.begin
}

func (h *MsgRequest) ReqLen() uint32 {
	return h.reqlen
}

// ============================================================================
// IMPL =======================================================================

func (h *MsgRequest) Encode() []byte {
	pl := h.msgBase.Encode()
	binary.BigEndian.PutUint32(pl, h.index)
	binary.BigEndian.PutUint32(pl, h.begin)
	binary.BigEndian.PutUint32(pl, h.reqlen)
	return pl
}

func (mp *MsgRequest) String() string {
	strb := strings.Builder{}
	strb.WriteString(mp.msgBase.String())
	strb.WriteString(fmt.Sprintf("Index: %v\n", mp.index))
	strb.WriteString(fmt.Sprintf("Begin: %v\n", mp.begin))
	strb.WriteString(fmt.Sprintf("Req Len: %v\n", mp.reqlen))
	return strb.String()
}

// ============================================================================
// FUNC =======================================================================

func DecodeMsgRequest(payload []byte) (*MsgRequest, error) {
	if uint32(len(payload)) != LenRequestPayload {
		return nil, fmt.Errorf("request message must have %v byte payload, got %v", LenRequestPayload, len(payload))
	}
	index := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	reqlen := binary.BigEndian.Uint32(payload[8:12])
	return NewMsgRequest(index, begin, reqlen), nil
}
