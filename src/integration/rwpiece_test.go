package integration

import (
	"bytes"
	"testing"

	"gotor/rw"
	"gotor/torrent/info"
	"gotor/utils"
	"gotor/utils/test"
)

func TestReadPiece(t *testing.T) {

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
			rwer := rw.NewReaderWriter(torInfo.Files())
			e = rwer.OCAT()
			defer func() {
				err := rwer.CloseAll()
				test.CheckError(t, err)
			}()
			test.CheckFatal(t, e)

			// Check all the pieces
			buf := make([]byte, tt.piecelen, tt.piecelen)
			for i, p := range pieces {

				plocs, e := torInfo.PieceLookup(int64(i))
				test.CheckError(t, e)

				n, e := rwer.ReadReqs(plocs, buf)
				test.CheckError(t, e)

				if !bytes.Equal(p, buf[:n]) {
					t.Fatalf("Piece(%v)\nWant: %v\n Got: %v", i, p, buf[:n])
				}
			}
		})
	}
}
