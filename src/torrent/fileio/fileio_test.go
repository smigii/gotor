package fileio

import (
	"bytes"
	"testing"

	"gotor/torrent/filesd"
	"gotor/torrent/info"
	"gotor/utils"
	"gotor/utils/test"
)

func TestFileIO_ReadPiece(t *testing.T) {

	type testFile struct {
		fpath string
		data  []byte
	}

	GetPaths := func(tfiles []testFile) []string {
		fpaths := make([]string, 0, len(tfiles))
		for _, tfile := range tfiles {
			fpaths = append(fpaths, tfile.fpath)
		}
		return fpaths
	}

	tests := []struct {
		name       string
		piecelen   int64
		workingDir string
		torName    string
		tfiles     []testFile
	}{
		{
			name:       "",
			workingDir: ".",
			piecelen:   3,
			torName:    "multifile",
			tfiles: []testFile{
				{"f1", []byte{'a', 'b', 'c', 'd'}},
				{"f2", []byte{'e'}},
				{"f3", []byte{'f', 'g', 'h', 'i', 'j', 'k', 'l'}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, tfile := range tt.tfiles {
					e := test.CleanUpTestFile(tfile.fpath)
					test.CheckFatal(t, e)
				}
			}()

			// Write the test files
			for _, tfile := range tt.tfiles {
				e := test.WriteTestFile(tfile.fpath, tfile.data)
				test.CheckFatal(t, e)
			}

			// Merge all data into a single slice
			data := make([]byte, 0)
			for _, tfile := range tt.tfiles {
				data = append(data, tfile.data...)
			}

			// Get the known pieces
			pieces := utils.SegmentData(data, tt.piecelen)

			// Create TorInfo
			fpaths := GetPaths(tt.tfiles)
			torInfo, e := info.CreateTorInfo(fpaths, tt.workingDir, tt.name, tt.piecelen)
			test.CheckError(t, e)

			// OCAT files, defer cleanup
			fileio := NewFileIO(torInfo)
			e = fileio.OCATAll(torInfo.Files())
			defer func() {
				err := fileio.CloseAll()
				test.CheckError(t, err)
			}()
			test.CheckFatal(t, e)

			// Check all the pieces
			buf := make([]byte, tt.piecelen, tt.piecelen)
			for i, p := range pieces {

				n, e := fileio.ReadPiece(int64(i), buf)
				test.CheckError(t, e)

				if !bytes.Equal(p, buf[:n]) {
					t.Fatalf("Piece(%v)\nWant: %v\n Got: %v", i, p, buf[:n])
				}
			}
		})
	}
}

func TestFileIO_WritePiece(t *testing.T) {

	type testFile struct {
		fpath string
		data  []byte
	}

	getFentries := func(tfiles []testFile) []filesd.EntryBase {
		fpaths := make([]filesd.EntryBase, 0, len(tfiles))
		for _, tfile := range tfiles {
			feb := filesd.MakeFileEntry(tfile.fpath, int64(len(tfile.data)))
			fpaths = append(fpaths, feb)
		}
		return fpaths
	}

	tests := []struct {
		name       string
		piecelen   int64
		workingDir string
		torName    string
		tfiles     []testFile
	}{
		{
			name:       "",
			workingDir: ".",
			piecelen:   3,
			torName:    "multifile",
			tfiles: []testFile{
				{"f1", []byte{'a', 'b', 'c', 'd'}},
				{"f2", []byte{'e'}},
				{"f3", []byte{'f', 'g', 'h', 'i', 'j', 'k', 'l'}},
			},
		},
		{
			name:       "",
			workingDir: ".",
			piecelen:   3,
			torName:    "singlefile",
			tfiles: []testFile{
				{"singlefile", []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				for _, tfile := range tt.tfiles {
					e := test.CleanUpTestFile(tfile.fpath)
					test.CheckFatal(t, e)
				}
			}()

			// Merge all data into a single slice
			data := make([]byte, 0)
			for _, tfile := range tt.tfiles {
				data = append(data, tfile.data...)
			}

			// Get the known pieces and their hashes
			pieces := utils.SegmentData(data, tt.piecelen)
			hashes := utils.HashSlices(pieces)

			// Make slice of file entry base
			entries := getFentries(tt.tfiles)

			// Create TorInfo
			torInfo, e := info.NewTorInfo(tt.torName, tt.piecelen, hashes, entries)
			test.CheckFatal(t, e)

			// OCAT files, defer cleanup
			fileio := NewFileIO(torInfo)
			e = fileio.OCATAll(torInfo.Files())
			defer func() {
				err := fileio.CloseAll()
				test.CheckError(t, err)
			}()
			test.CheckFatal(t, e)

			// Write all the pieces
			for i, piece := range pieces {
				_, e = fileio.WritePiece(int64(i), piece)
				test.CheckError(t, e)
			}

			// Check all the pieces
			buf := make([]byte, tt.piecelen, tt.piecelen)
			for i, piece := range pieces {

				n, e := fileio.ReadPiece(int64(i), buf)
				test.CheckError(t, e)

				if !bytes.Equal(piece, buf[:n]) {
					t.Fatalf("Piece(%v)\nWant: %v\n Got: %v", i, piece, buf[:n])
				}
			}
		})
	}
}
