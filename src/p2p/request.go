package p2p

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	// MsgRequestTotalLen is the total length of a request message (len 4 + id 1 + payload 12)
	MsgRequestTotalLen = uint8(17)

	// MsgRequestSpecLen is the length of the message as defined in BEP_0003
	MsgRequestSpecLen = uint32(13)

	// MsgRequestPayloadLen is the payload size in bytes
	MsgRequestPayloadLen = uint32(12)
)

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
			length: MsgRequestSpecLen,
			mtype:  TypeRequest,
		},
		index:  index,
		begin:  begin,
		reqlen: reqlen,
	}
}

// ============================================================================
// GETTER =====================================================================

func (mr *MsgRequest) Index() uint32 {
	return mr.index
}

func (mr *MsgRequest) Begin() uint32 {
	return mr.begin
}

func (mr *MsgRequest) ReqLen() uint32 {
	return mr.reqlen
}

// ============================================================================
// IMPL =======================================================================

func (mr *MsgRequest) Encode() []byte {
	pl := mr.msgBase.Encode()
	binary.BigEndian.PutUint32(pl, mr.index)
	binary.BigEndian.PutUint32(pl, mr.begin)
	binary.BigEndian.PutUint32(pl, mr.reqlen)
	return pl
}

func (mr *MsgRequest) String() string {
	strb := strings.Builder{}
	strb.WriteString("Message: Request\n")
	strb.WriteString(fmt.Sprintf("Index: %v\n", mr.index))
	strb.WriteString(fmt.Sprintf("Begin: %v\n", mr.begin))
	strb.WriteString(fmt.Sprintf("Req Len: %v", mr.reqlen))
	return strb.String()
}

// ============================================================================
// FUNC =======================================================================

func DecodeMsgRequest(payload []byte) (*MsgRequest, error) {
	if uint32(len(payload)) != MsgRequestPayloadLen {
		return nil, fmt.Errorf("request message must have %v byte payload, got %v", MsgRequestPayloadLen, len(payload))
	}
	index := binary.BigEndian.Uint32(payload[0:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	reqlen := binary.BigEndian.Uint32(payload[8:12])
	return NewMsgRequest(index, begin, reqlen), nil
}
