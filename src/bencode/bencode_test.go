package bencode

import (
	"bytes"
	"os"
	"testing"
)

func TestBencode(t *testing.T) {

	fdata, err := os.ReadFile("../../test_media/ubuntu-20.04.4-desktop-amd64.iso.torrent")
	if err != nil {
		t.Error(err)
	}

	d, err := Decode(fdata)
	if err != nil {
		t.Error(err)
	}

	dict, ok := d.(Dict)
	if !ok {
		t.Error("Error converting to dict")
	}

	if dict["announce"] != "https://torrent.ubuntu.com/announce" {
		t.Error("Bad 'announce' value")
	}

	e, err := Encode(d)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(fdata, e) {
		t.Error("encode(decode(fdata)) != fdata")
	}

}
