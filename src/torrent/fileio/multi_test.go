package fileio

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	fentry2 "gotor/torrent/filesd"
	"gotor/torrent/info"
	"gotor/utils"
	"gotor/utils/test"
)

func TestNewMultiFileHandler(t *testing.T) {

	type testStruct struct {
		fentry         fentry2.EntryBase
		wantStartPiece int64
		wantEndPiece   int64
		wantStartOff   int64
		wantEndOff     int64
	}

	makeTorFileList := func(structs []testStruct) []fentry2.EntryBase {
		l := make([]fentry2.EntryBase, 0, len(structs))
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
			{fentry2.MakeFileEntry("f1", 32), 0, 0, 0, 31}},
		},
		{"Multifile Simple", 5, 5, 25, []testStruct{
			{fentry2.MakeFileEntry("f1", 3), 0, 0, 0, 2},  // [0, 2]
			{fentry2.MakeFileEntry("f2", 5), 0, 1, 3, 2},  // [3, 7]
			{fentry2.MakeFileEntry("f3", 2), 1, 1, 3, 4},  // [8, 9]
			{fentry2.MakeFileEntry("f4", 13), 2, 4, 0, 2}, // [10, 22]
			{fentry2.MakeFileEntry("f5", 2), 4, 4, 3, 4}}, // [23, 24]
		},
		{"Multifile Truc", 5, 5, 22, []testStruct{
			{fentry2.MakeFileEntry("f1", 5), 0, 0, 0, 4},  // [0, 4]
			{fentry2.MakeFileEntry("f2", 5), 1, 1, 0, 4},  // [5, 9]
			{fentry2.MakeFileEntry("f3", 10), 2, 3, 0, 4}, // [10, 19]
			{fentry2.MakeFileEntry("f4", 2), 4, 4, 0, 1}}, // [20, 21]
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			files := makeTorFileList(tt.files)
			hashes := test.DummyHashes(tt.numPieces)
			testInfo, e := info.NewTorInfo("test", tt.pieceLen, hashes, files)
			test.CheckError(t, e)

			mfh := NewMultiFileHandler(testInfo)

			e = mfh.OCAT()
			defer func() {
				err := mfh.Close()
				test.CheckError(t, err)
				for _, tstruct := range tt.files {
					err = test.CleanUpTestFile(tstruct.fentry.LocalPath())
					test.CheckError(t, err)
				}
			}()
			test.CheckError(t, e)

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

	tests := []struct {
		name     string
		piecelen int64
		files    []fentry2.EntryBase
		kvp      map[int64][]string // Map piece index to file names that should be returned
	}{
		{"One Each", 3, []fentry2.EntryBase{
			fentry2.MakeFileEntry("f1", 3),
			fentry2.MakeFileEntry("f2", 3),
			fentry2.MakeFileEntry("f3", 3),
		}, map[int64][]string{
			0: {"f1"},
			1: {"f2"},
			2: {"f3"},
			3: {},
		}},
		{"Mix (up to 2)", 3, []fentry2.EntryBase{
			fentry2.MakeFileEntry("f1", 4),
			fentry2.MakeFileEntry("f2", 6),
			fentry2.MakeFileEntry("f3", 2),
		}, map[int64][]string{
			0: {"f1"},
			1: {"f1", "f2"},
			2: {"f2"},
			3: {"f2", "f3"},
			4: {},
		}},
		{"Triple", 9, []fentry2.EntryBase{
			fentry2.MakeFileEntry("f1", 3),
			fentry2.MakeFileEntry("f2", 3),
			fentry2.MakeFileEntry("f3", 3),
		}, map[int64][]string{
			0: {"f1", "f2", "f3"},
			1: {},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			numPieces := fentry2.CalcNumPieces(tt.files, tt.piecelen)
			hashes := test.DummyHashes(numPieces)
			testInfo, e := info.NewTorInfo("test", tt.piecelen, hashes, tt.files)
			test.CheckError(t, e)

			mfh := NewMultiFileHandler(testInfo)
			e = mfh.OCAT()
			defer func() {
				err := mfh.Close()
				test.CheckError(t, err)
				for _, fentry := range tt.files {
					err = test.CleanUpTestFile(fentry.LocalPath())
					test.CheckError(t, err)
				}
			}()
			test.CheckError(t, e)

			// Keys are piece indices, values are slices of file paths that
			// should be in there
			for k, v := range tt.kvp {

				files := mfh.GetFiles(k)

				got := make([]string, 0, len(files))
				for _, f := range files {
					got = append(got, f.TorPath())
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

	// Initialize some data
	dataLen := uint8(100)
	data := make([]byte, dataLen, dataLen)
	for i, _ := range data {
		data[i] = uint8(i)
	}

	tests := []struct {
		name     string
		piecelen int64
		files    []fentry2.EntryBase
	}{
		{"Tiny", 2, []fentry2.EntryBase{
			fentry2.MakeFileEntry("f1", 1),
		}},
		// [A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		{"Simple", 3, []fentry2.EntryBase{
			fentry2.MakeFileEntry("f1", 4),
			fentry2.MakeFileEntry("f2", 6),
			fentry2.MakeFileEntry("f3", 2),
		}},
		{"Multi", 4, []fentry2.EntryBase{
			fentry2.MakeFileEntry("f1", 5),
			fentry2.MakeFileEntry("f2", 1),
			fentry2.MakeFileEntry("f3", 1),
			fentry2.MakeFileEntry("f4", 1),
		}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, fentry := range tt.files {
					err := test.CleanUpTestFile(fentry.LocalPath())
					test.CheckError(t, err)
				}
			}()

			// Write the test files
			curs := int64(0) // Cursor for data byte slice
			for _, tf := range tt.files {
				e := test.WriteTestFile(tf.TorPath(), data[curs:curs+tf.Length()])
				if e != nil {
					t.Fatal(e)
				}

				curs += tf.Length()
			}

			// Create MultiFileHandler
			numPieces := fentry2.CalcNumPieces(tt.files, tt.piecelen)
			hashes := test.DummyHashes(numPieces)
			testInfo, e := info.NewTorInfo("test", tt.piecelen, hashes, tt.files)
			test.CheckError(t, e)

			mfh := NewMultiFileHandler(testInfo)
			e = mfh.OCAT()
			test.CheckError(t, e)
			defer func() {
				err := mfh.Close()
				test.CheckError(t, err)
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

	type TestFileEntry struct {
		fentry2.EntryBase
		data []byte
	}

	tests := []struct {
		name     string
		piecelen int64
		files    []TestFileEntry
	}{
		{"Tiny", 2, []TestFileEntry{
			{fentry2.MakeFileEntry("f1", 1), []byte{'a'}},
		}},
		// [A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		{"Simple", 3, []TestFileEntry{
			{fentry2.MakeFileEntry("f1", 4), []byte{'a', 'b', 'c', 'd'}},
			{fentry2.MakeFileEntry("f2", 6), []byte{'e', 'f', 'g', 'h', 'i', 'j'}},
			{fentry2.MakeFileEntry("f3", 2), []byte{'k', 'l'}},
		}},
		// [A|A|A|A]  [A|B|C|D]  [E|E| | ]
		{"Multi", 4, []TestFileEntry{
			{fentry2.MakeFileEntry("f1", 5), []byte{'a', 'b', 'c', 'd', 'e'}},
			{fentry2.MakeFileEntry("f2", 1), []byte{'f'}},
			{fentry2.MakeFileEntry("f3", 1), []byte{'g'}},
			{fentry2.MakeFileEntry("f4", 1), []byte{'h'}},
			{fentry2.MakeFileEntry("f5", 2), []byte{'i', 'j'}},
		}},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, tfe := range tt.files {
					err := test.CleanUpTestFile(tfe.LocalPath())
					test.CheckError(t, err)
				}
			}()

			// Create empty files and create a single data byte array
			data := make([]byte, 0, 0)
			fileEntries := make([]fentry2.EntryBase, 0, len(tt.files))
			for _, tf := range tt.files {
				f, e := utils.CreateZeroFilledFile(tf.TorPath(), tf.Length())
				test.CheckError(t, e)
				e = f.Close()
				test.CheckError(t, e)

				data = append(data, tf.data...)
				fileEntries = append(fileEntries, tf.EntryBase)
			}

			// Get the pieces and hashes
			pieces := utils.SegmentData(data, tt.piecelen)
			hashes := strings.Builder{}
			for _, p := range pieces {
				hashes.WriteString(utils.SHA1(p))
			}

			// Create MultiFileHandler
			testInfo, e := info.NewTorInfo("test", tt.piecelen, hashes.String(), fileEntries)
			test.CheckError(t, e)

			mfh := NewMultiFileHandler(testInfo)
			e = mfh.OCAT()
			test.CheckError(t, e)
			defer func() {
				err := mfh.Close()
				test.CheckError(t, err)
			}()

			// Write all the pieces
			for i, p := range pieces {
				e := mfh.Write(int64(i), p)
				test.CheckError(t, e)
			}

			// Read all the pieces
			got := make([]byte, tt.piecelen, tt.piecelen)
			for i, p := range pieces {
				n, e := mfh.Piece(int64(i), got)
				test.CheckError(t, e)

				if !bytes.Equal(got[:n], p) {
					t.Errorf("Piece(%v)\n Got: %v\nWant: %v", i, got, p)
				}
			}
		})
	}
}
