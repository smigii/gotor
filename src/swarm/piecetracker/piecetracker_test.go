package piecetracker

import (
	"testing"
)

func TestPeerPieceTracker_NextPiece(t *testing.T) {

	MakeBoolSlice := func(size int64, set ...int64) []bool {
		b := make([]bool, size, size)
		for _, i := range set {
			b[i] = true
		}
		return b
	}

	tests := []struct {
		name    string
		size    int64
		incs    []int64
		decs    []int64
		indices []int64
		want    int64
	}{
		// Simple case
		{"", 10, []int64{0}, []int64{}, []int64{0, 1, 2, 3}, 0},

		// Simple case
		{"", 10, []int64{0, 1, 0, 5, 1}, []int64{}, []int64{0, 1, 5}, 5},

		// Make sure decrementing works
		{"", 10, []int64{0, 0, 0, 1, 1, 1, 5, 5, 5}, []int64{0, 0, 1, 5}, []int64{0, 1, 5}, 0},

		// We aren't looking for the rarest piece
		{"", 10, []int64{0, 0, 1, 1, 2, 2, 2}, []int64{}, []int64{2}, 2},

		// There are non-zero pieces, but we don't have any of them
		{"", 10, []int64{0, 1, 2}, []int64{}, []int64{4}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppt := NewPeerPieceTracker(tt.size)

			for _, i := range tt.incs {
				ppt.IncPiece(i)
			}

			for _, d := range tt.decs {
				ppt.DecPiece(d)
			}

			// This will simulate our PeerHandler bitfield
			boolSlice := MakeBoolSlice(tt.size, tt.indices...)

			nextIdx := ppt.NextPiece(boolSlice)

			if nextIdx != tt.want {
				t.Errorf("NextPiece got %v, want %v", nextIdx, tt.want)
			}
		})
	}
}
