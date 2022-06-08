package p2p

import (
	"bytes"
	"testing"
)

func TestNoPayloadDecode(t *testing.T) {
	tests := []struct {
		name   string
		mtype  uint8
		length uint32
		data   []byte
		err    bool
	}{
		{
			name:   "Good Choke",
			data:   []byte{0, 0, 0, 1, 0},
			mtype:  TypeChoke,
			length: 1,
			err:    false,
		},
		{
			name:   "Good Unchoke",
			data:   []byte{0, 0, 0, 1, 1},
			mtype:  TypeUnchoke,
			length: 1,
			err:    false,
		},
		{
			name:   "Good Interested",
			data:   []byte{0, 0, 0, 1, 2},
			mtype:  TypeInterested,
			length: 1,
			err:    false,
		},
		{
			name:   "Good Uninterested",
			data:   []byte{0, 0, 0, 1, 3},
			mtype:  TypeNotInterested,
			length: 1,
			err:    false,
		},
		{
			name:   "Bad Choke",
			data:   []byte{0, 0, 0, 3, 0},
			mtype:  TypeChoke,
			length: 1,
			err:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := Decode(tt.data)
			if tt.err {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if tt.mtype != msg.Mtype() {
					t.Errorf("expected type %v, got %v", tt.mtype, msg.Mtype())
				}
				if tt.length != msg.Length() {
					t.Errorf("expected length %v, got %v", tt.length, msg.Length())
				}
			}
		})
	}
}

func TestNoPayloadEncode(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want []byte
	}{
		{
			name: "Encode Choke",
			msg:  NewMsgChoke(),
			want: []byte{0, 0, 0, 1, 0},
		},
		{
			name: "Encode Unchoke",
			msg:  NewMsgUnchoke(),
			want: []byte{0, 0, 0, 1, 1},
		},
		{
			name: "Encode Interested",
			msg:  NewMsgInterested(),
			want: []byte{0, 0, 0, 1, 2},
		},
		{
			name: "Encode Not Interested",
			msg:  NewMsgNotInterested(),
			want: []byte{0, 0, 0, 1, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := tt.msg.Encode()
			if !bytes.Equal(tt.want, enc) {
				t.Errorf("\nwant %v\n got %v", tt.want, enc)
			}
		})
	}
}
