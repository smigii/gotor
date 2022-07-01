package torrent

import (
	"encoding/hex"
	"testing"

	"gotor/utils"
)

func TestFromTorrentFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		infohash string
	}{
		{
			name:     "multifile",
			path:     "../../test/multifile.torrent",
			infohash: "a976fdd2ccce699eab604115408ead8560c2d095",
		},
		{
			name:     "medfile",
			path:     "../../test/medfile.torrent",
			infohash: "a554b7cc4616ae2fce43fabe9a7fe931aff5d85c",
		},
		{
			name:     "bigfile",
			path:     "../../test/bigfile.torrent",
			infohash: "658a9fbffae3c07e2320b7c4b144da978296dcc7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tor, err := FromTorrentFile(tt.path, ".")
			utils.CheckError(t, err)

			trueHashBytes, _ := hex.DecodeString(tt.infohash)
			trueHashString := string(trueHashBytes)

			if tor.Infohash() != trueHashString {
				t.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", tt.infohash, hex.EncodeToString([]byte(tor.Infohash())))
			}
		})
	}
}

func TestNewTorrent(t *testing.T) {
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
			workingDir: "../../test",
			piecelen:   32768,
			infohash:   "a976fdd2ccce699eab604115408ead8560c2d095",
			torName:    "multifile",
			paths: []string{
				"multifile/d1/d1.1/d1.1.file1",
				"multifile/d1/d1.2/d1.2.file1",
				"multifile/d2/d2.file1",
				"multifile/d2/d2.file2",
				"multifile/file1",
				"multifile/file2",
				"multifile/file3",
				"multifile/file4",
			},
		},
		{
			name:       "medfile",
			workingDir: "../../test",
			piecelen:   65536,
			infohash:   "a554b7cc4616ae2fce43fabe9a7fe931aff5d85c",
			torName:    "medfile",
			paths: []string{
				"medfile",
			},
		},
		{
			name:       "bigfile",
			workingDir: "../../test",
			piecelen:   1048576,
			infohash:   "658a9fbffae3c07e2320b7c4b144da978296dcc7",
			torName:    "bigfile",
			paths: []string{
				"../../test/bigfile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//tor, e := NewTorrent(tt.paths, tt.workingDir, tt.torName, "somewebsite.com/announce", tt.piecelen)
			//utils.CheckError(t, e)
			//
			//trueHashBytes, _ := hex.DecodeString(tt.infohash)
			//trueHashString := string(trueHashBytes)
			//
			//if tor.Infohash() != trueHashString {
			//	t.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", tt.infohash, hex.EncodeToString([]byte(tor.Infohash())))
			//}
		})
	}
}
