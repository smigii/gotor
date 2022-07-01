package fileio

import (
	"errors"
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

// TorInfo holds the metadata found in a torrent's info dictionary.
type TorInfo struct {
	name      string `info:"name"`
	pieceLen  int64  `info:"piece length"`
	hashes    string `info:"pieces"`
	numPieces int64
	length    int64       `info:"length"`
	files     []FileEntry `info:"files"`
	isSingle  bool        // Is this a single-file or multi-file torrent?
}

// ============================================================================
// GETTERS ====================================================================

func (t *TorInfo) Name() string {
	return t.name
}

func (t *TorInfo) PieceLen() int64 {
	return t.pieceLen
}

func (t *TorInfo) PieceHashes() string {
	return t.hashes
}

func (t *TorInfo) NumPieces() int64 {
	return t.numPieces
}

func (t *TorInfo) Length() int64 {
	return t.length
}

func (t *TorInfo) Files() []FileEntry {
	return t.files
}

func (t *TorInfo) IsSingle() bool {
	return t.isSingle
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewTorInfo(name string, pieceLen int64, hashes string, files []FileEntry) (*TorInfo, error) {

	if len(hashes)%20 != 0 {
		return nil, errors.New("hashes must be multiple of 20")
	}

	length := int64(0)
	for _, fentry := range files {
		length += fentry.Length()
	}

	return &TorInfo{
		name:      name,
		pieceLen:  pieceLen,
		hashes:    hashes,
		numPieces: int64(len(hashes) / 20),
		length:    length,
		files:     files,
		isSingle:  len(files) == 1,
	}, nil

}

// FromDict creates and returns a new *TorInfo from the info dictionary of a
// torrent file.
func FromDict(info bencode.Dict) (*TorInfo, error) {
	fdata := TorInfo{}
	var err error

	fdata.name, err = info.GetString("name")
	if err != nil {
		return nil, err
	}

	fdata.pieceLen, err = info.GetInt("piece length")
	if err != nil {
		return nil, err
	}

	fdata.hashes, err = info.GetString("pieces")
	if err != nil {
		return nil, err
	}
	if len(fdata.hashes)%20 != 0 {
		return nil, &FileMetaError{
			msg: fmt.Sprintf("'pieces' length must be multiple of 20, got length [%v]", len(fdata.hashes)),
		}
	}
	fdata.numPieces = int64(len(fdata.hashes) / 20)

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

func (t *TorInfo) PieceHash(idx int64) (string, error) {
	if idx >= t.numPieces {
		return "", &FileMetaError{
			msg: fmt.Sprintf("requested piece index [%v], max is [%v]", idx, t.numPieces-1),
		}
	}

	offset := idx * 20
	return t.hashes[offset : offset+20], nil
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
			torPath: strb.String()[:l-1], // exclude last '/'
			length:  fLen,
		})
	}

	return sfl, nil
}
