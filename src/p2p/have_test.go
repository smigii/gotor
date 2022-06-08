package p2p

import (
	"bytes"
	"testing"
)

func TestHaveDecode(t *testing.T) {
	tests := []struct {
		name string
		idx  uint32
		data []byte
		err  bool
	}{
		{
			name: "Have 1 byte index",
			idx:  12,
			data: []byte{0, 0, 0, 5, 4, 0, 0, 0, 12},
			err:  false,
		},
		{
			name: "Have 2 byte index",
			idx:  57005,
			data: []byte{0, 0, 0, 5, 4, 0x00, 0x00, 0xDE, 0xAD},
			err:  false,
		},
		{
			name: "Have 3 byte index",
			idx:  11189196,
			data: []byte{0, 0, 0, 5, 4, 0x00, 0xAA, 0xBB, 0xCC},
			err:  false,
		},
		{
			name: "Have 4 byte index",
			idx:  666420666,
			data: []byte{0, 0, 0, 5, 4, 0x27, 0xB8, 0xC5, 0xBA},
			err:  false,
		},
		{
			name: "Bad Have",
			idx:  12,
			data: []byte{0, 0, 0, 8, 4, 0, 0, 0, 12},
			err:  true,
		},
		{
			name: "Bad Have 2",
			idx:  12,
			data: []byte{0, 0, 0, 8, 4, 0, 0, 0, 12, 66, 66, 66, 66},
			err:  true,
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
				hmsg, ok := msg.(*MsgHave)
				if !ok {
					t.Error("couldn't convert message to HaveMsg")
				}
				if tt.idx != hmsg.Index() {
					t.Errorf("expected index %v, got %v", tt.idx, hmsg.Index())
				}
			}
		})
	}
}

func TestHaveEncode(t *testing.T) {
	tests := []struct {
		name string
		msg  *MsgHave
		want []byte
	}{
		{
			name: "Have 1 byte index",
			msg:  NewMsgHave(12),
			want: []byte{0, 0, 0, 5, 4, 0, 0, 0, 12},
		},
		{
			name: "Have 2 byte index",
			msg:  NewMsgHave(57005),
			want: []byte{0, 0, 0, 5, 4, 0x00, 0x00, 0xDE, 0xAD},
		},
		{
			name: "Have 3 byte index",
			msg:  NewMsgHave(11189196),
			want: []byte{0, 0, 0, 5, 4, 0x00, 0xAA, 0xBB, 0xCC},
		},
		{
			name: "Have 4 byte index",
			msg:  NewMsgHave(666420666),
			want: []byte{0, 0, 0, 5, 4, 0x27, 0xB8, 0xC5, 0xBA},
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
