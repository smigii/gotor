package p2p

import (
	"reflect"
	"testing"
)

func TestMsgPiece_Encode(t *testing.T) {

	tests := []struct {
		name  string
		index uint32
		begin uint32
		block []byte
		want  []byte
	}{
		{
			"", 0, 20, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			[]byte{
				0, 0, 0, (9 + 10), // len
				7,          // type
				0, 0, 0, 0, // index
				0, 0, 0, 20, // begin
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, // data
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := NewMsgPiece(tt.index, tt.begin, tt.block)
			got := mp.Encode()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode()\n got: %v\nwant: %v", got, tt.want)
			}
		})
	}
}
