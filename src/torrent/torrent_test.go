package torrent

import (
	"encoding/hex"
	"testing"

	"gotor/utils/test"
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
			test.CheckError(t, err)

			trueHashBytes, _ := hex.DecodeString(tt.infohash)
			trueHashString := string(trueHashBytes)

			if tor.Infohash() != trueHashString {
				t.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", tt.infohash, hex.EncodeToString([]byte(tor.Infohash())))
			}
		})
	}
}
