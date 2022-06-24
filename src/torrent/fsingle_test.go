package torrent

import (
	"bytes"
	"testing"

	"gotor/utils"
)

func TestFileSingle_Piece(t *testing.T) {

	testMeta := TorFileMeta{}
	testMeta.isSingle = true

	tests := []struct {
		name     string
		fpath    string
		piecelen int64
		data     []byte
	}{
		{"No trunc piece", "f1", 3, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}},
		{"Trunc piece", "f2", 3, []byte{0, 1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer utils.CleanUpTestFile(tt.fpath)

			pieces := utils.SegmentData(tt.data, tt.piecelen)

			testMeta.pieceLen = tt.piecelen
			testMeta.name = tt.fpath
			testMeta.numPieces = int64(len(pieces))
			testMeta.length = int64(len(tt.data))

			e := utils.WriteTestFile(tt.fpath, tt.data)
			if e != nil {
				t.Fatal(e)
			}

			f, e := newFileSingle(&testMeta)
			if e != nil {
				t.Error(e)
			}

			for i := 0; i < len(pieces); i++ {
				got, err := f.Piece(int64(i))
				if err != nil {
					t.Error(e)
				}

				if !bytes.Equal(got, pieces[i]) {
					t.Errorf("Piece(%v)\n Got: %v\nWant: %v", i, got, pieces[i])
				}
			}
		})
	}
}
