package torrent

import (
	"fmt"
	"gotor/bencode"
	"strings"
)

// ============================================================================
// STRUCTS ====================================================================

// TorFileMeta holds the relevent metadata of the files in torrent.
type TorFileMeta struct {
	name      string
	pieceLen  uint64
	pieces    string
	numPieces uint64
	length    uint64
	files     []torFileEntry
	isSingle  bool // Is this a single-file or multi-file torrent?
}

// torFileEntry holds only the data contained in a torrent's info dictionary.
// This exists to facilitate testing; rather than passing in a bencode.List to
// testing setup, we can pass these instead.
type torFileEntry struct {
	length uint64
	fpath  string
}

// ============================================================================
// GETTERS ====================================================================

func (t *TorFileMeta) Name() string {
	return t.name
}

func (t *TorFileMeta) PieceLen() uint64 {
	return t.pieceLen
}

func (t *TorFileMeta) Pieces() string {
	return t.pieces
}

func (t *TorFileMeta) NumPieces() uint64 {
	return t.numPieces
}

func (t *TorFileMeta) Length() uint64 {
	return t.length
}

func (t *TorFileMeta) Files() []torFileEntry {
	return t.files
}

func (t *TorFileMeta) IsSingle() bool {
	return t.isSingle
}

// ============================================================================
// CONSTRUCTOR ================================================================

func newTorFileMeta(info bencode.Dict) (*TorFileMeta, error) {
	fdata := TorFileMeta{}
	var err error

	fdata.name, err = info.GetString("name")
	if err != nil {
		return nil, err
	}

	fdata.pieceLen, err = info.GetUint("piece length")
	if err != nil {
		return nil, err
	}

	fdata.pieces, err = info.GetString("pieces")
	if err != nil {
		return nil, err
	}
	if len(fdata.pieces)%20 != 0 {
		return nil, &TorError{
			msg: fmt.Sprintf("'pieces' length must be multiple of 20, got length [%v]", len(fdata.pieces)),
		}
	}
	fdata.numPieces = uint64(len(fdata.pieces) / 20)

	fdata.length, err = info.GetUint("length")
	if err == nil {
		fdata.isSingle = true
		// TODO: Finish
	} else {
		fdata.isSingle = false

		// Try 'files'
		files, err := info.GetList("files")
		if err != nil {
			return nil, &TorError{
				msg: fmt.Sprintf("missing keys 'length' and 'files', must have exactly 1"),
			}
		}

		// Read through list of file dictionaries
		fdata.files, err = extractFileEntries(files, fdata.name)
		if err != nil {
			return nil, err
		}
	}

	return &fdata, nil
}

// ============================================================================
// FUNC =======================================================================

func (t *TorFileMeta) Piece(idx uint64) (string, error) {
	if idx >= t.numPieces {
		return "", &TorError{
			msg: fmt.Sprintf("requested piece index [%v], max is [%v]", idx, t.numPieces-1),
		}
	}

	offset := idx * 20
	return t.pieces[offset:20], nil
}

// extractFileEntries extracts the {path, length} dictionaries from a bencoded
// list.
func extractFileEntries(benlist bencode.List, dirname string) ([]torFileEntry, error) {
	sfl := make([]torFileEntry, 0, 4)

	for _, fEntry := range benlist {
		fDict, ok := fEntry.(bencode.Dict)
		if !ok {
			return nil, &TorError{
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
		strb := strings.Builder{}

		// Write the directory name
		strb.WriteString(dirname)
		strb.WriteByte('/')

		for _, fPathEntry := range fPathList {
			pathPiece, ok := fPathEntry.(string)
			if !ok {
				return nil, &TorError{
					msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
				}
			}
			strb.WriteString(pathPiece)
			strb.WriteByte('/')
		}
		l := len(strb.String())

		sfl = append(sfl, torFileEntry{
			fpath:  strb.String()[:l-1], // exclude last '/'
			length: fLen,
		})
	}

	return sfl, nil
}
