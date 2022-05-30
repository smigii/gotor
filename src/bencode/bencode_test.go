package bencode

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestBencode(t *testing.T) {

	tests := []string{
		"../../test_media/multifile.torrent",
		"../../test_media/medfile.torrent",
		"../../test_media/bigfile.torrent",
	}

	for _, f := range tests {
		err := testTorrent(f)
		if err != nil {
			t.Error(err)
		}
	}

}

func testTorrent(path string) error {

	fdata, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	d, err := Decode(fdata)
	if err != nil {
		return err
	}

	e, err := Encode(d)
	if err != nil {
		return err
	}

	if !bytes.Equal(fdata, e) {
		return errors.New("encode(decode(fdata)) != fdata")
	}

	return nil
}
