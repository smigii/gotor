package fileio

import (
	"errors"
	"fmt"
	"os"

	"gotor/bf"
	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// SingleFileHandler is used for single-file torrents. It can access pieces faster
// than MultiFileHandler since it doesn't need to find out which files are contained
// in a given piece.
type SingleFileHandler struct {
	meta *TorFileMeta
	fp   *os.File
	bf   *bf.Bitfield
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewSingleFileHandler(meta *TorFileMeta) (*SingleFileHandler, error) {
	fp, err := utils.OpenCheck(meta.name, meta.length)
	if err != nil {
		return nil, err
	} else {
		fs := SingleFileHandler{
			meta: meta,
			fp:   fp,
			bf:   bf.NewBitfield(meta.NumPieces()),
		}
		//err = fs.Validate()  // broken
		return &fs, err
	}
}

// ============================================================================
// IMPL =======================================================================

func (f *SingleFileHandler) Piece(index int64) ([]byte, error) {
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

	_, e := f.fp.ReadAt(buf, seekAmnt)

	return buf, e
}

func (f *SingleFileHandler) Write(index int64, data []byte) error {
	meta := f.meta
	if index >= meta.numPieces {
		return fmt.Errorf("index out of bounds, got %v, max %v", index, meta.numPieces)
	}

	hash := utils.SHA1(data)
	knownHash, _ := meta.PieceHash(index)

	if hash != knownHash {
		return &HashError{}
	}

	seekAmnt := index * meta.pieceLen

	_, e := f.fp.WriteAt(data, seekAmnt)

	return e
}

func (f *SingleFileHandler) Validate() error {
	var i int64

	for i = 0; i < f.meta.numPieces; i++ {

		knownHash, e := f.meta.PieceHash(i)
		if e != nil {
			return e
		}

		// TODO: Grabbing pieces one at a time is slow
		piece, e := f.Piece(i)
		if e != nil {
			return e
		}

		val := utils.SHA1(piece) == knownHash
		f.bf.Set(i, val)
	}

	return nil
}

func (f *SingleFileHandler) FileMeta() *TorFileMeta {
	return f.meta
}

func (f *SingleFileHandler) Bitfield() *bf.Bitfield {
	return f.bf
}

func (f *SingleFileHandler) Close() error {
	if f.fp != nil {
		return f.fp.Close()
	} else {
		return errors.New("nil file pointer")
	}
}
