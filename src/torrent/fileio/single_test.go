package fileio

import (
	"bytes"
	"os"
	"testing"

	fentry2 "gotor/torrent/filesd"
	"gotor/torrent/info"
	"gotor/utils"
	"gotor/utils/test"
)

func TestFileSingle_Piece(t *testing.T) {

	tests := []struct {
		name     string
		fpath    string
		piecelen int64
		data     []byte
	}{
		{"No trunc piece", "single1", 3, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}},
		{"Trunc piece", "single2", 3, []byte{0, 1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := test.WriteTestFile(tt.fpath, tt.data)
			defer func() {
				err := test.CleanUpTestFile(tt.fpath)
				test.CheckError(t, err)
			}()
			test.CheckFatal(t, e)

			pieces := utils.SegmentData(tt.data, tt.piecelen)
			hashes := utils.HashSlices(pieces)

			fentry := []fentry2.EntryBase{fentry2.MakeFileEntry(tt.fpath, int64(len(tt.data)))}
			testInfo, e := info.NewTorInfo("test", tt.piecelen, hashes, fentry)

			sfh := NewSingleFileHandler(testInfo)
			e = sfh.OCAT()
			defer func() {
				err := sfh.Close()
				test.CheckError(t, err)
			}()
			test.CheckError(t, e)

			got := make([]byte, tt.piecelen, tt.piecelen)
			for i := 0; i < len(pieces); i++ {
				n, err := sfh.Piece(int64(i), got)
				test.CheckError(t, err)

				if !bytes.Equal(got[:n], pieces[i]) {
					t.Errorf("Piece(%v)\n Got: %v\nWant: %v", i, got, pieces[i])
				}
			}
		})
	}
}

func TestFileSingle_Write(t *testing.T) {

	tests := []struct {
		name     string
		fpath    string
		pieceLen int64
		data     []byte
	}{
		{"One Piece", "test1", 3, []byte{0, 1, 2}},
		{"Two Pieces", "test1", 3, []byte{0, 1, 2, 3, 4}},
		{"Three Pieces", "test3", 3, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				e := test.CleanUpTestFile(tt.fpath)
				test.CheckError(t, e)
			}()

			pieces := utils.SegmentData(tt.data, tt.pieceLen)
			hashes := utils.HashSlices(pieces)

			fentry := []fentry2.EntryBase{fentry2.MakeFileEntry(tt.fpath, int64(len(tt.data)))}
			testInfo, e := info.NewTorInfo("test", tt.pieceLen, hashes, fentry)

			sfh := NewSingleFileHandler(testInfo)
			e = sfh.OCAT()
			defer func() {
				err := sfh.Close()
				test.CheckError(t, err)
			}()
			test.CheckError(t, e)

			for i := 0; i < len(pieces); i++ {
				e = sfh.Write(int64(i), pieces[i])
				test.CheckError(t, e)
			}

			fdata, e := os.ReadFile(tt.fpath)
			test.CheckError(t, e)

			if !bytes.Equal(tt.data, fdata) {
				t.Errorf("Error writing data\nWant: %v\nRead: %v", tt.data, fdata)
			}
		})
	}
}
