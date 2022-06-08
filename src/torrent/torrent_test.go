package torrent

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestTorrent(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		infohash string
	}{
		{
			name:     "multifile",
			path:     "../../test/multifile.torrent",
			infohash: "b253474bd8536994e4ea9e0786a4e0ea528e2530",
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
			err := testTorrent(tt.path, tt.infohash)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func testTorrent(path string, infohash string) error {
	tor, err := NewTorrent(path)

	if err != nil {
		return err
	}

	trueHashBytes, _ := hex.DecodeString(infohash)
	trueHashString := string(trueHashBytes)

	if tor.Infohash() != trueHashString {
		return fmt.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", infohash, hex.EncodeToString([]byte(tor.Infohash())))
	}

	return nil
}
