package torrent

import (
	"errors"
	"fmt"
	"os"

	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// FileSingle is used for single-file torrents. It can access pieces faster
// than FileList since it doesn't need to find out which files are contained
// in a given piece.
type FileSingle struct {
	meta *TorFileMeta
	fp   *os.File
	bf   *utils.Bitfield
}

// ============================================================================
// CONSTRUCTOR ================================================================

func newFileSingle(meta *TorFileMeta) (*FileSingle, error) {
	fp, err := utils.OpenCheck(meta.name, meta.length)
	if err != nil {
		return nil, err
	} else {
		fs := FileSingle{
			meta: meta,
			fp:   fp,
			bf:   utils.NewBitfield(meta.NumPieces()),
		}
		//err = fs.Validate()  // broken
		return &fs, err
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

	_, e := f.fp.ReadAt(buf, seekAmnt)

	return buf, e
}

func (f *FileSingle) Write(index int64, data []byte) error {
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

func (f *FileSingle) Validate() error {
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

func (f *FileSingle) FileMeta() *TorFileMeta {
	return f.meta
}

func (f *FileSingle) Bitfield() *utils.Bitfield {
	return f.bf
}

func (f *FileSingle) Close() error {
	if f.fp != nil {
		return f.fp.Close()
	} else {
		return errors.New("nil file pointer")
	}
}
