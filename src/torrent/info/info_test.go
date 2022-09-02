package info

import (
	"encoding/hex"
	"testing"

	"gotor/bencode"
	"gotor/torrent/filesd"
	"gotor/utils"
	"gotor/utils/test"
)

func TestCreateTorInfo(t *testing.T) {
	tests := []struct {
		name         string
		piecelen     int64
		infohash     string
		workingDir   string
		torName      string
		lastPieceLen int64
		paths        []string
	}{
		{
			name:         "multifile",
			workingDir:   "../../../test",
			piecelen:     32768,
			infohash:     "a976fdd2ccce699eab604115408ead8560c2d095",
			torName:      "multifile",
			lastPieceLen: 28416,
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
			name:         "medfile",
			workingDir:   "../../../test",
			piecelen:     65536,
			infohash:     "a554b7cc4616ae2fce43fabe9a7fe931aff5d85c",
			torName:      "medfile",
			lastPieceLen: 65536,
			paths: []string{
				"medfile",
			},
		},
		{
			name:         "bigfile",
			workingDir:   "../../../test",
			piecelen:     1048576,
			infohash:     "658a9fbffae3c07e2320b7c4b144da978296dcc7",
			torName:      "bigfile",
			lastPieceLen: 1048576,
			paths: []string{
				"bigfile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			info, e := CreateTorInfo(tt.paths, tt.workingDir, tt.name, tt.piecelen)
			test.CheckError(t, e)

			bcoded := info.Bencode()
			enc, e := bencode.Encode(bcoded)
			infohash := utils.SHA1(enc)

			trueHashBytes, _ := hex.DecodeString(tt.infohash)
			trueHashString := string(trueHashBytes)

			if infohash != trueHashString {
				t.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", tt.infohash, hex.EncodeToString([]byte(infohash)))
			}

			if info.LastPieceLen() != tt.lastPieceLen {
				t.Errorf("bad lastPieceLen\nwant %v\n got %v", tt.lastPieceLen, info.LastPieceLen())
			}
		})
	}
}

func TestTorInfo_LastPieceLen(t *testing.T) {
	type fields struct {
		name         string
		pieceLen     int64
		hashes       string
		numPieces    int64
		length       int64
		flist        filesd.FileList
		isSingle     bool
		pm           PieceMap
		lastPieceLen int64
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &TorInfo{
				name:         tt.fields.name,
				pieceLen:     tt.fields.pieceLen,
				hashes:       tt.fields.hashes,
				numPieces:    tt.fields.numPieces,
				length:       tt.fields.length,
				flist:        tt.fields.flist,
				isSingle:     tt.fields.isSingle,
				pm:           tt.fields.pm,
				lastPieceLen: tt.fields.lastPieceLen,
			}
			if got := ti.LastPieceLen(); got != tt.want {
				t.Errorf("LastPieceLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
