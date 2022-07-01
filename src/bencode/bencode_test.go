package bencode

import (
	"bytes"
	"os"
	"testing"

	"gotor/utils/test"
)

func TestBencode(t *testing.T) {

	tests := []struct {
		name string
		path string
	}{
		{"multi file", "../../test/multifile.torrent"},
		{"medium file", "../../test/medfile.torrent"},
		{"big file", "../../test/bigfile.torrent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fdata, err := os.ReadFile(tt.path)
			test.CheckError(t, err)

			d, err := Decode(fdata)
			test.CheckError(t, err)

			e, err := Encode(d)
			test.CheckError(t, err)

			if !bytes.Equal(fdata, e) {
				t.Error("encode(decode(fdata)) != fdata")
			}
		})
	}

}
