package p2p

import (
	"bytes"
	"testing"
)

func TestBitfieldDecode(t *testing.T) {
	tests := []struct {
		name string
		len  uint32
		data []byte
		bf   []byte
		err  bool
	}{
		{
			name: "Good bitfield (short)",
			len:  5,
			data: []byte{0, 0, 0, 5, 5, 0, 1, 2, 3},
			err:  false,
		},
		{
			name: "Bad bitfield",
			len:  8,
			data: []byte{0, 0, 0, 8, 5, 1, 2, 3, 4, 5, 6},
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
				if err != nil {
					t.Error(err)
				}
				bfmsg, ok := msg.(*MsgBitfield)
				if !ok {
					t.Error("couldn't convert to MsgBitfield")
				}
				want := tt.data[PayloadStart:]
				if !bytes.Equal(want, bfmsg.Bitfield()) {
					t.Errorf("\nwant %v\n got %v", want, bfmsg.Bitfield())
				}
			}
		})
	}
}

func TestBitfieldEncode(t *testing.T) {
	tests := []struct {
		name string
		bf   []byte
		base []byte
	}{
		{
			name: "bitfield",
			bf:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			base: []byte{0, 0, 0, 11, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := append(tt.base, tt.bf...)
			msg, err := NewMsgBitfield(tt.bf, uint32(len(tt.bf))+1)
			if err != nil {
				t.Error(err)
			}
			enc := msg.Encode()
			if !bytes.Equal(want, enc) {
				t.Errorf("\nwant %v\n got %v", want, enc)
			}
		})
	}
}
