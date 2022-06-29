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

func (f *SingleFileHandler) Piece(index int64, buf []byte) (int64, error) {
	meta := f.meta

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

	e := f.rw.Read(f.entry.fpath, seekAmnt, buf[:readAmnt])
	return readAmnt, e
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

	e := f.rw.Write(f.entry.fpath, seekAmnt, data)

	return e
}

func (f *SingleFileHandler) Validate() error {
	//var i int64

	//for i = 0; i < f.meta.numPieces; i++ {
	//
	//	knownHash, e := f.meta.PieceHash(i)
	//	if e != nil {
	//		return e
	//	}
	//
	//	// TODO: Grabbing pieces one at a time is slow
	//	piece, e := f.Piece(i)
	//	if e != nil {
	//		return e
	//	}
	//
	//	val := utils.SHA1(piece) == knownHash
	//	f.bf.Set(i, val)
	//}

	return nil
}

func (f *SingleFileHandler) FileMeta() *TorFileMeta {
	return f.meta
}

func (f *SingleFileHandler) Bitfield() *bf.Bitfield {
	return f.bf
}

func (f *SingleFileHandler) Close() error {
	return f.rw.CloseAll()
}
