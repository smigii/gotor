package torrent

import (
	"gotor/bencode"
	"testing"
)

type testStruct struct {
	path           string
	len            int64
	wantStartPiece uint64
	wantEndPiece   uint64
	wantStartOff   uint64
	wantEndOff     uint64
}

func testEncode(structs []testStruct) bencode.List {
	var l bencode.List
	for _, s := range structs {
		d := bencode.Dict{
			"path":   bencode.List{s.path},
			"length": s.len,
		}
		l = append(l, d)
	}
	return l
}

func checkField(t *testing.T, name string, expect uint64, got uint64) {
	if expect != got {
		t.Errorf("%v expected %v, got %v", name, expect, got)
	}
}

func TestNewFileList(t *testing.T) {

	testTor := Torrent{}

	tests := []struct {
		name      string
		pieceLen  uint64
		numPieces uint64
		files     []testStruct
	}{
		{"Single File", 32, 1, []testStruct{
			{"f1", 32, 0, 0, 0, 31}},
		},
		{"Multifile Simple", 5, 5, []testStruct{
			{"f1", 3, 0, 0, 0, 2},  // [0, 2]
			{"f2", 5, 0, 1, 3, 2},  // [3, 7]
			{"f3", 2, 1, 1, 3, 4},  // [8, 9]
			{"f4", 13, 2, 4, 0, 2}, // [10, 22]
			{"f5", 2, 4, 4, 3, 4}}, // [23, 24]
		},
		{"Multifile Truc", 5, 5, []testStruct{
			{"f1", 5, 0, 0, 0, 4},  // [0, 4]
			{"f2", 5, 1, 1, 0, 4},  // [5, 9]
			{"f3", 10, 2, 3, 0, 4}, // [10, 19]
			{"f4", 2, 4, 4, 0, 1}}, // [20, 21]
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testTor.pieceLen = tt.pieceLen
			testTor.numPieces = tt.numPieces

			benlist := testEncode(tt.files)
			flist, err := newFileList(benlist, &testTor)
			if err != nil {
				t.Error(err)
			}

			for i, fe := range flist.Files() {
				checkField(t, "Start Piece", tt.files[i].wantStartPiece, fe.StartPiece())
				checkField(t, "End Piece", tt.files[i].wantEndPiece, fe.EndPiece())
				checkField(t, "Start Piece Offset", tt.files[i].wantStartOff, fe.StartPieceOff())
				checkField(t, "End Piece Offset", tt.files[i].wantEndOff, fe.EndPieceOff())
			}
		})
	}
}
