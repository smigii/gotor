package torrent

import (
	"crypto/sha1"
	"fmt"
	"os"

	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// FileSingle is used for single-file torrents. It can access pieces faster
// than FileList since it doesn't need to find out which files are contained
// int a given piece.
type FileSingle struct {
	meta *TorFileMeta
	fp   *os.File
}

// ============================================================================
// CONSTRUCTOR ================================================================

func newFileSingle(meta *TorFileMeta) (*FileSingle, error) {
	fp, err := utils.OpenCheck(meta.name, int64(meta.length))
	if err != nil {
		return nil, err
	} else {
		return &FileSingle{
			meta: meta,
			fp:   fp,
		}, nil
	}
}

// ============================================================================
// IMPL =======================================================================

func (f *FileSingle) Piece(index int64) ([]byte, error) {
	meta := f.meta

	if index >= meta.numPieces {
		return nil, fmt.Errorf("attempted to get index %v, max is %v", index, meta.numPieces-1)
	}

	seekAmnt := meta.pieceLen * index
	readAmnt := meta.pieceLen
	truncAmnt := meta.length % meta.pieceLen

	// Last piece may be truncated
	if index == meta.numPieces-1 && truncAmnt != 0 {
		readAmnt = truncAmnt
	}

	buf := make([]byte, readAmnt, readAmnt)
	_, e := f.fp.ReadAt(buf, int64(seekAmnt))
	return buf, e
}

func (f *FileSingle) Write(index int64, data []byte) error {
	meta := f.meta
	if index >= meta.numPieces {
		return fmt.Errorf("index out of bounds, got %v, max %v", index, meta.numPieces)
	}

	hasher := sha1.New()
	hasher.Write(data)
	hash := string(hasher.Sum(nil))
	wantHash, _ := meta.PieceHash(index)

	if hash != wantHash {
		return &HashError{}
	}

	seekAmnt := index * meta.pieceLen
	_, e := f.fp.WriteAt(data, int64(seekAmnt))
	return e
}

func (f *FileSingle) FileMeta() *TorFileMeta {
	return f.meta
}

func (f *FileSingle) Validate(bf *utils.Bitfield) {
	//TODO implement me
}
