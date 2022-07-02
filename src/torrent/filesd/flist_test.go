package filesd

import (
	"testing"
)

func TestMakeFileList(t *testing.T) {

	makeTorFileList := func(entries []Entry) []EntryBase {
		l := make([]EntryBase, 0, len(entries))
		for _, v := range entries {
			l = append(l, v.EntryBase)
		}
		return l
	}

	checkField := func(t *testing.T, name string, expect int64, got int64) {
		if expect != got {
			t.Errorf("%v expected %v, got %v", name, expect, got)
		}
	}

	tests := []struct {
		name      string
		pieceLen  int64
		numPieces int64
		totalLen  int64
		files     []Entry
	}{
		{"Single File", 32, 1, 32, []Entry{
			{MakeFileEntry("f1", 32), 0, 0, 0, 31}},
		},
		{"Multifile Simple", 5, 5, 25, []Entry{
			{MakeFileEntry("f1", 3), 0, 0, 0, 2},  // [0, 2]
			{MakeFileEntry("f2", 5), 0, 1, 3, 2},  // [3, 7]
			{MakeFileEntry("f3", 2), 1, 1, 3, 4},  // [8, 9]
			{MakeFileEntry("f4", 13), 2, 4, 0, 2}, // [10, 22]
			{MakeFileEntry("f5", 2), 4, 4, 3, 4}}, // [23, 24]
		},
		{"Multifile Truc", 5, 5, 22, []Entry{
			{MakeFileEntry("f1", 5), 0, 0, 0, 4},  // [0, 4]
			{MakeFileEntry("f2", 5), 1, 1, 0, 4},  // [5, 9]
			{MakeFileEntry("f3", 10), 2, 3, 0, 4}, // [10, 19]
			{MakeFileEntry("f4", 2), 4, 4, 0, 1}}, // [20, 21]
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			files := makeTorFileList(tt.files)
			flist := MakeFileList(files, tt.pieceLen)

			for i, fe := range flist {
				checkField(t, "Start Piece", tt.files[i].StartPiece(), fe.StartPiece())
				checkField(t, "End Piece", tt.files[i].EndPiece(), fe.EndPiece())
				checkField(t, "Start Piece Offset", tt.files[i].StartPieceOff(), fe.StartPieceOff())
				checkField(t, "End Piece Offset", tt.files[i].EndPieceOff(), fe.EndPieceOff())
			}
		})
	}
}
