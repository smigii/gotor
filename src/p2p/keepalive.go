/* keepalive.go ===============================================================
BEP_0003 defines keep alives to be [0, 0, 0, 0], which makes it kinda annoying
to return a message interface when it doesn't have a type like the rest of the
messages.
============================================================================ */

package p2p

const MsgKeepAliveTotalLen = uint8(4)

type MsgKeepAlive struct {
	msgBase
}

// KeepAliveSingleton is a global keep alive message. We will be sending lots of these,
// doesn't make sense to allocate a new one every time.
var KeepAliveSingleton = MsgKeepAlive{
	msgBase: msgBase{
		length: 0,
		mtype:  TypeKeepAlive,
	},
}

func (ka *MsgKeepAlive) Encode() []byte {
	return []byte{0, 0, 0, 0}
}

func (ka *MsgKeepAlive) String() string {
	return "Message: Keep Alive"
}
