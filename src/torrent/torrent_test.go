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
			path:     "../../test_media/flowers.torrent",
			infohash: "9ac55b9c736b1f97d510d7c53c7b6210421cbd06",
		},
		{
			path:     "../../test_media/ubuntu-20.04.4-desktop-amd64.iso.torrent",
			infohash: "f09c8d0884590088f4004e010a928f8b6178c2fd",
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
