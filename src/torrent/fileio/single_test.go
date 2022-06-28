package fileio

import (
	"bytes"
	"os"
	"strings"
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
			defer func() {
				err := utils.CleanUpTestFile(tt.fpath)
				utils.CheckError(t, err)
			}()

			pieces := utils.SegmentData(tt.data, tt.piecelen)

			testMeta.pieceLen = tt.piecelen
			testMeta.name = tt.fpath
			testMeta.numPieces = int64(len(pieces))
			testMeta.length = int64(len(tt.data))

			e := utils.WriteTestFile(tt.fpath, tt.data)
			utils.CheckFatal(t, e)

			fs, e := NewSingleFileHandler(&testMeta)
			defer func() {
				err := fs.Close()
				utils.CheckError(t, err)
			}()
			utils.CheckError(t, e)

			for i := 0; i < len(pieces); i++ {
				got, err := fs.Piece(int64(i))
				utils.CheckError(t, err)

				if !bytes.Equal(got, pieces[i]) {
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
				e := utils.CleanUpTestFile(tt.fpath)
				utils.CheckError(t, e)
			}()

			pieces := utils.SegmentData(tt.data, tt.pieceLen)
			hashes := strings.Builder{}
			for _, p := range pieces {
				hashes.WriteString(utils.SHA1(p))
			}

			testMeta := TorFileMeta{
				name:      tt.fpath,
				pieceLen:  tt.pieceLen,
				pieces:    hashes.String(),
				numPieces: int64(len(pieces)),
				length:    int64(len(tt.data)),
				files:     nil,
				isSingle:  true,
			}

			fs, e := NewSingleFileHandler(&testMeta)
			defer func() {
				err := fs.Close()
				utils.CheckError(t, err)
			}()
			utils.CheckError(t, e)

			for i := 0; i < len(pieces); i++ {
				e = fs.Write(int64(i), pieces[i])
				utils.CheckError(t, e)
			}

			fdata, e := os.ReadFile(tt.fpath)
			utils.CheckError(t, e)

			if !bytes.Equal(tt.data, fdata) {
				t.Errorf("Error writing data\nWant: %v\nRead: %v", tt.data, fdata)
			}
		})
	}
}