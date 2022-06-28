package fileio

import (
	"fmt"
	"strings"

	"gotor/bencode"
)

// ============================================================================
// ERRORS =====================================================================

type FileMetaError struct{ msg string }

func (fme *FileMetaError) Error() string {
	return "tracker error: " + fme.msg
}

// ============================================================================
// STRUCTS ====================================================================

// TorFileMeta holds the relevent metadata of the files in torrent.
type TorFileMeta struct {
	name      string
	pieceLen  int64
	pieces    string
	numPieces int64
	length    int64
	files     []FileEntry
	isSingle  bool // Is this a single-file or multi-file torrent?
}

// ============================================================================
// GETTERS ====================================================================

func (t *TorFileMeta) Name() string {
	return t.name
}

func (t *TorFileMeta) PieceLen() int64 {
	return t.pieceLen
}

func (t *TorFileMeta) PieceHashes() string {
	return t.pieces
}

func (t *TorFileMeta) NumPieces() int64 {
	return t.numPieces
}

func (t *TorFileMeta) Length() int64 {
	return t.length
}

func (t *TorFileMeta) Files() []FileEntry {
	return t.files
}

func (t *TorFileMeta) IsSingle() bool {
	return t.isSingle
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewFileMeta(info bencode.Dict) (*TorFileMeta, error) {
	fdata := TorFileMeta{}
	var err error

	fdata.name, err = info.GetString("name")
	if err != nil {
		return nil, err
	}

	fdata.pieceLen, err = info.GetInt("piece length")
	if err != nil {
		return nil, err
	}

	fdata.pieces, err = info.GetString("pieces")
	if err != nil {
		return nil, err
	}
	if len(fdata.pieces)%20 != 0 {
		return nil, &FileMetaError{
			msg: fmt.Sprintf("'pieces' length must be multiple of 20, got length [%v]", len(fdata.pieces)),
		}
	}
	fdata.numPieces = int64(len(fdata.pieces) / 20)

	fdata.length, err = info.GetInt("length")
	if err == nil {
		fdata.isSingle = true
		// TODO: Finish
	} else {
		fdata.isSingle = false

		// Try 'files'
		files, err := info.GetList("files")
		if err != nil {
			return nil, &FileMetaError{
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

func (t *TorFileMeta) PieceHash(idx int64) (string, error) {
	if idx >= t.numPieces {
		return "", &FileMetaError{
			msg: fmt.Sprintf("requested piece index [%v], max is [%v]", idx, t.numPieces-1),
		}
	}

	offset := idx * 20
	return t.pieces[offset : offset+20], nil
}

// extractFileEntries extracts the {path, length} dictionaries from a bencoded
// list.
func extractFileEntries(benlist bencode.List, dirname string) ([]FileEntry, error) {
	sfl := make([]FileEntry, 0, 4)

	for _, fEntry := range benlist {
		fDict, ok := fEntry.(bencode.Dict)
		if !ok {
			return nil, &FileMetaError{
				msg: fmt.Sprintf("failed to convert file entry to dictionary\n%v", fEntry),
			}
		}

		fLen, err := fDict.GetInt("length")
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
				return nil, &FileMetaError{
					msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
				}
			}
			strb.WriteString(pathPiece)
			strb.WriteByte('/')
		}
		l := len(strb.String())

		sfl = append(sfl, FileEntry{
			fpath:  strb.String()[:l-1], // exclude last '/'
			length: fLen,
		})
	}

	return sfl, nil
}