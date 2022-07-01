package fileio

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"gotor/utils"
)

func TestNewMultiFileHandler(t *testing.T) {

	testFileMeta := TorInfo{}

	type testStruct struct {
		fentry         FileEntry
		wantStartPiece int64
		wantEndPiece   int64
		wantStartOff   int64
		wantEndOff     int64
	}

	makeTorFileList := func(structs []testStruct) []FileEntry {
		l := make([]FileEntry, 0, len(structs))
		for _, v := range structs {
			l = append(l, v.fentry)
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
			{MakeFileEntry("f1", 32), 0, 0, 0, 31}},
		},
		{"Multifile Simple", 5, 5, 25, []testStruct{
			{MakeFileEntry("f1", 3), 0, 0, 0, 2},  // [0, 2]
			{MakeFileEntry("f2", 5), 0, 1, 3, 2},  // [3, 7]
			{MakeFileEntry("f3", 2), 1, 1, 3, 4},  // [8, 9]
			{MakeFileEntry("f4", 13), 2, 4, 0, 2}, // [10, 22]
			{MakeFileEntry("f5", 2), 4, 4, 3, 4}}, // [23, 24]
		},
		{"Multifile Truc", 5, 5, 22, []testStruct{
			{MakeFileEntry("f1", 5), 0, 0, 0, 4},  // [0, 4]
			{MakeFileEntry("f2", 5), 1, 1, 0, 4},  // [5, 9]
			{MakeFileEntry("f3", 10), 2, 3, 0, 4}, // [10, 19]
			{MakeFileEntry("f4", 2), 4, 4, 0, 1}}, // [20, 21]
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sfl := makeTorFileList(tt.files)
			testFileMeta.files = sfl
			testFileMeta.pieceLen = tt.pieceLen
			testFileMeta.numPieces = tt.numPieces

			mfh := NewMultiFileHandler(&testFileMeta)

			e := mfh.OCAT()
			defer func() {
				err := mfh.Close()
				utils.CheckError(t, err)
				for _, tstruct := range tt.files {
					err = utils.CleanUpTestFile(tstruct.fentry.LocalPath())
					utils.CheckError(t, err)
				}
			}()
			utils.CheckError(t, e)

			checkField(t, "Total Length", tt.totalLen, mfh.TorInfo().Length())
			for i, fe := range mfh.Files() {
				checkField(t, "Start Piece", tt.files[i].wantStartPiece, fe.StartPiece())
				checkField(t, "End Piece", tt.files[i].wantEndPiece, fe.EndPiece())
				checkField(t, "Start Piece Offset", tt.files[i].wantStartOff, fe.StartPieceOff())
				checkField(t, "End Piece Offset", tt.files[i].wantEndOff, fe.EndPieceOff())
			}
		})
	}
}

func TestGetFiles(t *testing.T) {

	testFileMeta := TorInfo{}

	tests := []struct {
		name     string
		piecelen int64
		files    []FileEntry
		kvp      map[int64][]string // Map piece index to file names that should be returned
	}{
		{"One Each", 3, []FileEntry{
			MakeFileEntry("f1", 3),
			MakeFileEntry("f2", 3),
			MakeFileEntry("f3", 3),
		}, map[int64][]string{
			0: {"f1"},
			1: {"f2"},
			2: {"f3"},
			3: {},
		}},
		{"Mix (up to 2)", 3, []FileEntry{
			MakeFileEntry("f1", 4),
			MakeFileEntry("f2", 6),
			MakeFileEntry("f3", 2),
		}, map[int64][]string{
			0: {"f1"},
			1: {"f1", "f2"},
			2: {"f2"},
			3: {"f2", "f3"},
			4: {},
		}},
		{"Triple", 9, []FileEntry{
			MakeFileEntry("f1", 3),
			MakeFileEntry("f2", 3),
			MakeFileEntry("f3", 3),
		}, map[int64][]string{
			0: {"f1", "f2", "f3"},
			1: {},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testFileMeta.files = tt.files
			testFileMeta.pieceLen = tt.piecelen

			mfh := NewMultiFileHandler(&testFileMeta)
			e := mfh.OCAT()
			defer func() {
				err := mfh.Close()
				utils.CheckError(t, err)
				for _, fentry := range tt.files {
					err = utils.CleanUpTestFile(fentry.LocalPath())
					utils.CheckError(t, err)
				}
			}()
			utils.CheckError(t, e)

			// Keys are piece indices, values are slices of file paths that
			// should be in there
			for k, v := range tt.kvp {

				files := mfh.GetFiles(k)

				got := make([]string, 0, len(files))
				for _, f := range files {
					got = append(got, f.torPath)
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

func TestMultiFileHandler_Piece(t *testing.T) {

	testFileMeta := TorInfo{}

	// Initialize some data
	dataLen := uint8(100)
	data := make([]byte, dataLen, dataLen)
	for i, _ := range data {
		data[i] = uint8(i)
	}

	tests := []struct {
		name     string
		piecelen int64
		files    []FileEntry
	}{
		{"Tiny", 2, []FileEntry{
			MakeFileEntry("f1", 1),
		}},
		// [A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		{"Simple", 3, []FileEntry{
			MakeFileEntry("f1", 4),
			MakeFileEntry("f2", 6),
			MakeFileEntry("f3", 2),
		}},
		{"Multi", 4, []FileEntry{
			MakeFileEntry("f1", 5),
			MakeFileEntry("f2", 1),
			MakeFileEntry("f3", 1),
			MakeFileEntry("f4", 1),
		}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, fentry := range tt.files {
					err := utils.CleanUpTestFile(fentry.LocalPath())
					utils.CheckError(t, err)
				}
			}()

			// Write the test files
			curs := int64(0) // Cursor for data byte slice
			for _, tf := range tt.files {
				e := utils.WriteTestFile(tf.torPath, data[curs:curs+tf.length])
				if e != nil {
					t.Fatal(e)
				}

				curs += tf.length
			}

			// Create MultiFileHandler
			testFileMeta.files = tt.files
			testFileMeta.pieceLen = tt.piecelen

			mfh := NewMultiFileHandler(&testFileMeta)
			e := mfh.OCAT()
			utils.CheckError(t, e)
			defer func() {
				err := mfh.Close()
				utils.CheckError(t, err)
			}()

			// Loop through all pieces and verify a match
			pieces := utils.SegmentData(data[:curs], tt.piecelen)
			npieces := int64(len(pieces))
			got := make([]byte, tt.piecelen, tt.piecelen)

			var i int64
			for i = 0; i < npieces; i++ {
				expect := pieces[i]
				n, err := mfh.Piece(i, got)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(expect, got[:n]) {
					t.Fatalf("Piece(%v)\nWant: %v\n Got: %v", i, expect, got)
				}
			}
		})
	}
}

func TestMultiFileHandler_Write(t *testing.T) {
	testFileMeta := TorInfo{}

	type TestFileEntry struct {
		FileEntry
		data []byte
	}

	tests := []struct {
		name     string
		piecelen int64
		files    []TestFileEntry
	}{
		{"Tiny", 2, []TestFileEntry{
			{MakeFileEntry("f1", 1), []byte{'a'}},
		}},
		// [A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		{"Simple", 3, []TestFileEntry{
			{MakeFileEntry("f1", 4), []byte{'a', 'b', 'c', 'd'}},
			{MakeFileEntry("f2", 6), []byte{'e', 'f', 'g', 'h', 'i', 'j'}},
			{MakeFileEntry("f3", 2), []byte{'k', 'l'}},
		}},
		// [A|A|A|A]  [A|B|C|D]  [E|E| | ]
		{"Multi", 4, []TestFileEntry{
			{MakeFileEntry("f1", 5), []byte{'a', 'b', 'c', 'd', 'e'}},
			{MakeFileEntry("f2", 1), []byte{'f'}},
			{MakeFileEntry("f3", 1), []byte{'g'}},
			{MakeFileEntry("f4", 1), []byte{'h'}},
			{MakeFileEntry("f5", 2), []byte{'i', 'j'}},
		}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, tfe := range tt.files {
					err := utils.CleanUpTestFile(tfe.LocalPath())
					utils.CheckError(t, err)
				}
			}()

			// Create empty files and create a single data byte array
			data := make([]byte, 0, 0)
			fileEntries := make([]FileEntry, 0, len(tt.files))
			for _, tf := range tt.files {
				f, e := utils.CreateZeroFilledFile(tf.torPath, tf.length)
				utils.CheckError(t, e)
				e = f.Close()
				utils.CheckError(t, e)

				data = append(data, tf.data...)
				fileEntries = append(fileEntries, tf.FileEntry)
			}

			// Get the pieces and hashes
			pieces := utils.SegmentData(data, tt.piecelen)
			hashes := strings.Builder{}
			for _, p := range pieces {
				hashes.WriteString(utils.SHA1(p))
			}

			// Create MultiFileHandler
			testFileMeta.hashes = hashes.String()
			testFileMeta.files = fileEntries
			testFileMeta.pieceLen = tt.piecelen
			testFileMeta.length = int64(len(data))
			testFileMeta.isSingle = false
			testFileMeta.numPieces = int64(len(pieces))
			testFileMeta.name = "testwrite"

			mfh := NewMultiFileHandler(&testFileMeta)
			e := mfh.OCAT()
			utils.CheckError(t, e)
			defer func() {
				err := mfh.Close()
				utils.CheckError(t, err)
			}()

			// Write all the pieces
			for i, p := range pieces {
				e := mfh.Write(int64(i), p)
				utils.CheckError(t, e)
			}

			// Read all the pieces
			got := make([]byte, tt.piecelen, tt.piecelen)
			for i, p := range pieces {
				n, e := mfh.Piece(int64(i), got)
				utils.CheckError(t, e)

				if !bytes.Equal(got[:n], p) {
					t.Errorf("Piece(%v)\n Got: %v\nWant: %v", i, got, p)
				}
			}
		})
	}
}
