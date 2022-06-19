package p2p

import (
	"bytes"
	"testing"
)

func TestKeepAliveDecode(t *testing.T) {
	data := []byte{0, 0, 0, 0}
	dr, err := Decode(data)
	msg := dr.Msg

	if err != nil {
		t.Error(err)
	}

	if msg.Length() != 0 {
		t.Errorf("expected length 0, got %v", msg.Length())
	}

	if msg.Mtype() != TypeKeepAlive {
		t.Errorf("expected type keep alive (%v), got %v", TypeKeepAlive, msg.Mtype())
	}
}

func TestKeepAliveEncode(t *testing.T) {
	msg := KeepAliveSingleton
	enc := msg.Encode()
	if !bytes.Equal(enc, []byte{0, 0, 0, 0}) {
		t.Errorf("expected [0,0,0,0], got %v", enc)
	}
}
