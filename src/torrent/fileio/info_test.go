package fileio

import (
	"encoding/hex"
	"testing"

	"gotor/bencode"
	"gotor/utils"
)

func TestCreateTorInfo(t *testing.T) {
	tests := []struct {
		name       string
		piecelen   int64
		infohash   string
		workingDir string
		torName    string
		paths      []string
	}{
		{
			name:       "multifile",
			workingDir: "../../../test",
			piecelen:   32768,
			infohash:   "a976fdd2ccce699eab604115408ead8560c2d095",
			torName:    "multifile",
			paths: []string{
				"d1/d1.1/d1.1.file1",
				"d1/d1.2/d1.2.file1",
				"d2/d2.file1",
				"d2/d2.file2",
				"file1",
				"file2",
				"file3",
				"file4",
			},
		},
		{
			name:       "medfile",
			workingDir: "../../../test",
			piecelen:   65536,
			infohash:   "a554b7cc4616ae2fce43fabe9a7fe931aff5d85c",
			torName:    "medfile",
			paths: []string{
				"medfile",
			},
		},
		{
			name:       "bigfile",
			workingDir: "../../../test",
			piecelen:   1048576,
			infohash:   "658a9fbffae3c07e2320b7c4b144da978296dcc7",
			torName:    "bigfile",
			paths: []string{
				"bigfile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			info, e := CreateTorInfo(tt.paths, tt.workingDir, tt.name, tt.piecelen)
			utils.CheckError(t, e)

			bcoded := info.Bencode()
			enc, e := bencode.Encode(bcoded)
			infohash := utils.SHA1(enc)

			trueHashBytes, _ := hex.DecodeString(tt.infohash)
			trueHashString := string(trueHashBytes)

			if infohash != trueHashString {
				t.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", tt.infohash, hex.EncodeToString([]byte(infohash)))
			}
		})
	}
}
