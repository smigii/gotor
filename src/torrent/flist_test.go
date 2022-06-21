package torrent

import (
	"reflect"
	"testing"
)

func TestNewFileList(t *testing.T) {

	testTor := Torrent{}

	type testStruct struct {
		tfe            torFileEntry
		wantStartPiece uint64
		wantEndPiece   uint64
		wantStartOff   uint64
		wantEndOff     uint64
	}

	mkSimpleFileList := func(structs []testStruct) []torFileEntry {
		l := make([]torFileEntry, 0, len(structs))
		for _, v := range structs {
			l = append(l, v.tfe)
		}
		return l
	}

	checkField := func(t *testing.T, name string, expect uint64, got uint64) {
		if expect != got {
			t.Errorf("%v expected %v, got %v", name, expect, got)
		}
	}

	tests := []struct {
		name      string
		pieceLen  uint64
		numPieces uint64
		files     []testStruct
	}{
		{"Single File", 32, 1, []testStruct{
			{torFileEntry{32, "f1"}, 0, 0, 0, 31}},
		},
		{"Multifile Simple", 5, 5, []testStruct{
			{torFileEntry{3, "f1"}, 0, 0, 0, 2},  // [0, 2]
			{torFileEntry{5, "f2"}, 0, 1, 3, 2},  // [3, 7]
			{torFileEntry{2, "f3"}, 1, 1, 3, 4},  // [8, 9]
			{torFileEntry{13, "f4"}, 2, 4, 0, 2}, // [10, 22]
			{torFileEntry{2, "f5"}, 4, 4, 3, 4}}, // [23, 24]
		},
		{"Multifile Truc", 5, 5, []testStruct{
			{torFileEntry{5, "f1"}, 0, 0, 0, 4},  // [0, 4]
			{torFileEntry{5, "f2"}, 1, 1, 0, 4},  // [5, 9]
			{torFileEntry{10, "f3"}, 2, 3, 0, 4}, // [10, 19]
			{torFileEntry{2, "f4"}, 4, 4, 0, 1}}, // [20, 21]
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testTor.pieceLen = tt.pieceLen
			testTor.numPieces = tt.numPieces

			sfl := mkSimpleFileList(tt.files)
			flist := newFileList(sfl, tt.pieceLen)

			for i, fe := range flist.Files() {
				checkField(t, "Start Piece", tt.files[i].wantStartPiece, fe.StartPiece())
				checkField(t, "End Piece", tt.files[i].wantEndPiece, fe.EndPiece())
				checkField(t, "Start Piece Offset", tt.files[i].wantStartOff, fe.StartPieceOff())
				checkField(t, "End Piece Offset", tt.files[i].wantEndOff, fe.EndPieceOff())
			}
		})
	}
}

func TestGetFiles(t *testing.T) {

	tests := []struct {
		name     string
		piecelen uint64
		files    []torFileEntry
		kvp      map[uint64][]string // Map piece index to file names that should be returned
	}{
		{"", 3, []torFileEntry{
			{4, "f1"},
			{6, "f2"},
			{2, "f3"},
		}, map[uint64][]string{
			0: {"f1"},
			1: {"f1", "f2"},
			2: {"f2"},
			3: {"f2", "f3"},
			4: {},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fl := newFileList(tt.files, tt.piecelen)

			// Keys are piece indices, values are slices of file paths that
			// should be in there
			for k, v := range tt.kvp {

				files := fl.GetFiles(k)

				got := make([]string, 0, len(files))
				for _, f := range files {
					got = append(got, f.fpath)
				}

				if len(files) != len(v) {
					t.Errorf("GetFiles(%v)\n got: %v\nwant: %v", k, got, v)
				}

				if !reflect.DeepEqual(got, v) {
					t.Errorf("GetFiles(%v)\n got: %v\nwant: %v", k, got, v)
				}
			}

		})
	}
}
