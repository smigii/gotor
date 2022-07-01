package fileio

import (
	"fmt"
	"io"

	"gotor/bf"
	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// SingleFileHandler is used for single-file torrents. It can access pieces faster
// than MultiFileHandler since it doesn't need to find out which files are contained
// in a given piece.
type SingleFileHandler struct {
	meta  *TorInfo
	entry *FileEntry
	rw    *readerWriter
	bf    *bf.Bitfield
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewSingleFileHandler(meta *TorInfo) *SingleFileHandler {
	fentry := FileEntry{
		length:  meta.Length(),
		torPath: meta.Name(),
	}
	rw := NewReaderWriter([]FileEntry{fentry})

	return &SingleFileHandler{
		meta:  meta,
		rw:    rw,
		entry: &fentry,
		bf:    bf.NewBitfield(meta.NumPieces()),
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

	_, e := sfh.rw.Read(sfh.entry.torPath, seekAmnt, buf[:readAmnt])
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

	e := sfh.rw.Write(sfh.entry.torPath, seekAmnt, data)

	return e
}

func (sfh *SingleFileHandler) Validate() error {

	const mib = 1048576
	const maxBuf = 10 * mib

	meta := sfh.FileMeta()
	npieces := maxBuf / meta.PieceLen()
	bufSize := npieces * meta.PieceLen()

	buf := make([]byte, bufSize, bufSize)
	seek := int64(0)
	index := int64(0) // Piece index

	for {
		n, e := sfh.rw.Read(meta.Name(), seek, buf)
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		seek += n

		pieces := utils.SegmentData(buf, meta.PieceLen())
		for _, piece := range pieces {
			wantHash, _ := meta.PieceHash(index)
			gotHash := utils.SHA1(piece)
			eq := wantHash == gotHash
			sfh.Bitfield().Set(index, eq)
			index++
		}
	}

	return nil
}

func (sfh *SingleFileHandler) FileMeta() *TorInfo {
	return sfh.meta
}

func (sfh *SingleFileHandler) Bitfield() *bf.Bitfield {
	return sfh.bf
}

func (sfh *SingleFileHandler) OCAT() error {
	e := sfh.rw.OCAT()
	return e

	//e = sfh.Validate()
	//return e
}

func (sfh *SingleFileHandler) Close() error {
	return sfh.rw.CloseAll()
}
