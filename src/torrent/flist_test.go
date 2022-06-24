package torrent

import (
	"bytes"
	"reflect"
	"testing"

	"gotor/utils"
)

func TestNewFileList(t *testing.T) {

	testFileMeta := TorFileMeta{}

	type testStruct struct {
		tfe            torFileEntry
		wantStartPiece int64
		wantEndPiece   int64
		wantStartOff   int64
		wantEndOff     int64
	}

	makeTorFileList := func(structs []testStruct) []torFileEntry {
		l := make([]torFileEntry, 0, len(structs))
		for _, v := range structs {
			l = append(l, v.tfe)
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
		files     []testStruct
	}{
		{"Single File", 32, 1, 32, []testStruct{
			{torFileEntry{32, "f1"}, 0, 0, 0, 31}},
		},
		{"Multifile Simple", 5, 5, 25, []testStruct{
			{torFileEntry{3, "f1"}, 0, 0, 0, 2},  // [0, 2]
			{torFileEntry{5, "f2"}, 0, 1, 3, 2},  // [3, 7]
			{torFileEntry{2, "f3"}, 1, 1, 3, 4},  // [8, 9]
			{torFileEntry{13, "f4"}, 2, 4, 0, 2}, // [10, 22]
			{torFileEntry{2, "f5"}, 4, 4, 3, 4}}, // [23, 24]
		},
		{"Multifile Truc", 5, 5, 22, []testStruct{
			{torFileEntry{5, "f1"}, 0, 0, 0, 4},  // [0, 4]
			{torFileEntry{5, "f2"}, 1, 1, 0, 4},  // [5, 9]
			{torFileEntry{10, "f3"}, 2, 3, 0, 4}, // [10, 19]
			{torFileEntry{2, "f4"}, 4, 4, 0, 1}}, // [20, 21]
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sfl := makeTorFileList(tt.files)
			testFileMeta.files = sfl
			testFileMeta.pieceLen = tt.pieceLen
			testFileMeta.numPieces = tt.numPieces

			flist := newFileList(&testFileMeta)

			checkField(t, "Total Length", tt.totalLen, flist.FileMeta().Length())
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

	testFileMeta := TorFileMeta{}

	tests := []struct {
		name     string
		piecelen int64
		files    []torFileEntry
		kvp      map[int64][]string // Map piece index to file names that should be returned
	}{
		{"One Each", 3, []torFileEntry{
			{3, "f1"},
			{3, "f2"},
			{3, "f3"},
		}, map[int64][]string{
			0: {"f1"},
			1: {"f2"},
			2: {"f3"},
			3: {},
		}},
		{"Mix (up to 2)", 3, []torFileEntry{
			{4, "f1"},
			{6, "f2"},
			{2, "f3"},
		}, map[int64][]string{
			0: {"f1"},
			1: {"f1", "f2"},
			2: {"f2"},
			3: {"f2", "f3"},
			4: {},
		}},
		{"Triple", 9, []torFileEntry{
			{3, "f1"},
			{3, "f2"},
			{3, "f3"},
		}, map[int64][]string{
			0: {"f1", "f2", "f3"},
			1: {},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testFileMeta.files = tt.files
			testFileMeta.pieceLen = tt.piecelen
			fl := newFileList(&testFileMeta)

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

func TestPiece(t *testing.T) {

	testFileMeta := TorFileMeta{}

	// Initialize some data
	dataLen := uint8(100)
	data := make([]byte, dataLen, dataLen)
	for i, _ := range data {
		data[i] = uint8(i)
	}

	tests := []struct {
		name     string
		piecelen int64
		files    []torFileEntry
	}{
		{"Tiny", 2, []torFileEntry{
			{1, "A"},
		}},
		// [A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		{"Simple", 3, []torFileEntry{
			{4, "A"},
			{6, "B"},
			{2, "C"},
		}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, tf := range tt.files {
					utils.CleanUpTestFile(tf.fpath)
				}
			}()

			// Write the test files
			curs := int64(0) // Cursor for data byte slice
			for _, tf := range tt.files {
				e := utils.WriteTestFile(tf.fpath, data[curs:curs+tf.length])
				if e != nil {
					t.Fatal(e)
				}

				curs += tf.length
			}

			// Create FileList
			testFileMeta.files = tt.files
			testFileMeta.pieceLen = tt.piecelen
			fl := newFileList(&testFileMeta)

			// Loop through all pieces and verify a match
			pieces := utils.SegmentData(data[:curs], tt.piecelen)
			npieces := int64(len(pieces))

			var i int64
			for i = 0; i < npieces; i++ {
				expect := pieces[i]
				got, err := fl.Piece(i)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(expect, got) {
					t.Fatalf("Piece(%v)\nWant: %v\n Got: %v", i, expect, got)
				}
			}
		})
	}
}
