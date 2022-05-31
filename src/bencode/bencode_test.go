package bencode

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestBencode(t *testing.T) {

	tests := []struct {
		name string
		path string
	}{
		{"multi file", "../../test_media/multifile.torrent"},
		{"medium file", "../../test_media/medfile.torrent"},
		{"big file", "../../test_media/bigfile.torrent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := testTorrent(tt.path)
			if e != nil {
				t.Error(e)
			}
		})
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
