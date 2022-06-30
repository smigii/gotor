package fileio

import (
	"fmt"

	"gotor/bf"
	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// SingleFileHandler is used for single-file torrents. It can access pieces faster
// than MultiFileHandler since it doesn't need to find out which files are contained
// in a given piece.
type SingleFileHandler struct {
	meta  *TorFileMeta
	entry *FileEntry
	rw    *readerWriter
	bf    *bf.Bitfield
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewSingleFileHandler(meta *TorFileMeta) (*SingleFileHandler, error) {
	fentry := FileEntry{
		length: meta.Length(),
		fpath:  meta.Name(),
	}
	rw, err := NewReaderWriter([]FileEntry{fentry})
	if err != nil {
		return nil, err
	} else {
		fs := SingleFileHandler{
			meta:  meta,
			rw:    rw,
			entry: &fentry,
			bf:    bf.NewBitfield(meta.NumPieces()),
		}
		//err = fs.Validate()  // broken
		return &fs, err
	}
}

// ============================================================================
// IMPL =======================================================================

func (sfh *SingleFileHandler) Piece(index int64, buf []byte) (int64, error) {
	meta := sfh.meta

	if index >= meta.numPieces {
		return 0, fmt.Errorf("attempted to get index %v, max is %v", index, meta.numPieces-1)
	}

	seekAmnt := meta.pieceLen * index
	readAmnt := meta.pieceLen
	truncAmnt := meta.length % meta.pieceLen

	// Last piece may be truncated
	if index == meta.numPieces-1 && truncAmnt != 0 {
		readAmnt = truncAmnt
	}

	if int64(len(buf)) < readAmnt {
		return 0, fmt.Errorf("buffer to small, need %v, got %v", readAmnt, len(buf))
	}

	e := sfh.rw.Read(sfh.entry.fpath, seekAmnt, buf[:readAmnt])
	return readAmnt, e
}

func (sfh *SingleFileHandler) Write(index int64, data []byte) error {
	meta := sfh.meta
	if index >= meta.numPieces {
		return fmt.Errorf("index out of bounds, got %v, max %v", index, meta.numPieces)
	}

	hash := utils.SHA1(data)
	knownHash, _ := meta.PieceHash(index)

	if hash != knownHash {
		return &HashError{}
	}

	seekAmnt := index * meta.pieceLen

	e := sfh.rw.Write(sfh.entry.fpath, seekAmnt, data)

	return e
}

func (sfh *SingleFileHandler) Validate() error {

	return nil
}

func (sfh *SingleFileHandler) FileMeta() *TorFileMeta {
	return sfh.meta
}

func (sfh *SingleFileHandler) Bitfield() *bf.Bitfield {
	return sfh.bf
}

func (sfh *SingleFileHandler) Close() error {
	return sfh.rw.CloseAll()
}
