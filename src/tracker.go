package main

import (
	"crypto/sha1"
	"fmt"
	"gotor/bencode"
	"os"
)

type TorrentError struct{ msg string }

func (te *TorrentError) Error() string {
	return "tracker error: " + te.msg
}

type Torrent struct {
	infohash string
	announce string
	name     string
	pieceLen uint64
	pieces   string
	nPieces  uint64
	length   uint64
	files    []TorrentFileEntry
}

type TorrentFileEntry struct {
	length uint64
	path   []string
}

func NewTorrent(path string) (*Torrent, error) {

	tor := Torrent{}
	var err error

	fdata, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	d, err := bencode.Decode(fdata)
	if err != nil {
		return nil, err
	}

	dict, ok := d.(bencode.Dict)
	if !ok {
		return nil, &TorrentError{
			msg: "decoded bencoding is not a dictionary",
		}
	}

	tor.announce, err = dict.GetString("announce")
	if err != nil {
		return nil, err
	}

	info, err := dict.GetDict("info")
	if err != nil {
		return nil, err
	}

	tor.name, err = info.GetString("name")
	if err != nil {
		return nil, err
	}

	tor.pieceLen, err = info.GetUint("piece length")
	if err != nil {
		return nil, err
	}

	// TODO: Read info dictionary manually for SHA1
	// This is rather wasteful
	hasher := sha1.New()
	enc, _ := bencode.Encode(info)
	hasher.Write(enc)
	tor.infohash = string(hasher.Sum(nil))

	// Pieces
	tor.pieces, err = info.GetString("pieces")
	if err != nil {
		return nil, err
	}
	if len(tor.pieces)%20 != 0 {
		return nil, &TorrentError{
			msg: fmt.Sprintf("'pieces' length must be multiple of 20, got length [%v]", len(tor.pieces)),
		}
	}
	tor.nPieces = uint64(len(tor.pieces) / 20)

	// Length string XOR Files dictionary
	tor.length, err = info.GetUint("length")
	if err != nil {

		// Try 'files'
		files, err := info.GetList("files")
		if err != nil {
			return nil, &TorrentError{
				msg: fmt.Sprintf("missing keys 'length' and 'files', must have exactly 1"),
			}
		}

		// Read through list of file dictionaries
		tor.files = make([]TorrentFileEntry, 0, 8)
		for _, fEntry := range files {
			fDict, ok := fEntry.(bencode.Dict)
			if !ok {
				return nil, &TorrentError{
					msg: fmt.Sprintf("failed to convert file entry to dictionary\n%v", fEntry),
				}
			}
			fLen, err := fDict.GetUint("length")
			if err != nil {
				return nil, err
			}
			fPathList, err := fDict.GetList("path")
			if err != nil {
				return nil, err
			}

			// Read through list of path strings
			pathPieces := make([]string, 0, 2)
			for _, fPathEntry := range fPathList {
				pathPiece, ok := fPathEntry.(string)
				if !ok {
					return nil, &TorrentError{
						msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
					}
				}
				pathPieces = append(pathPieces, pathPiece)
			}

			tor.files = append(tor.files, TorrentFileEntry{
				length: fLen,
				path:   pathPieces,
			})
		}
	}

	return &tor, nil
}

func (tor *Torrent) GetPiece(idx uint64) (string, error) {
	if idx >= tor.nPieces {
		return "", &TorrentError{
			msg: fmt.Sprintf("requested piece index [%v], only have [%v]", idx, tor.nPieces),
		}
	}

	offset := idx * 20
	return tor.pieces[offset:20], nil
}