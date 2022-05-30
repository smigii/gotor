package torrent

import (
	"encoding/hex"
	"fmt"
	"testing"
)

type testEntry struct {
	path     string
	infohash string
}

func TestTorrent(t *testing.T) {

	tests := []testEntry{
		{
			path:     "../../test_media/multifile.torrent",
			infohash: "b253474bd8536994e4ea9e0786a4e0ea528e2530",
		},
		{
			path:     "../../test_media/medfile.torrent",
			infohash: "a554b7cc4616ae2fce43fabe9a7fe931aff5d85c",
		},
		{
			path:     "../../test_media/bigfile.torrent",
			infohash: "658a9fbffae3c07e2320b7c4b144da978296dcc7",
		},
	}

	for _, test := range tests {
		err := testTorrent(test.path, test.infohash)
		if err != nil {
			t.Error(err)
		}
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
		return fmt.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", infohash, hex.EncodeToString([]byte(tor.infohash)))
	}

	return nil
}
