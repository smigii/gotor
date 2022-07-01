package fileio

import (
	"fmt"
	"io"

	"gotor/bf"
	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// SingleFileHandler is used for single-file torrents. It can access pieces
// faster than MultiFileHandler since it doesn't need to find out which files
// are contained in a given piece.
type SingleFileHandler struct {
	info   *TorInfo
	fentry *FileEntry
	rw     *readerWriter
	bf     *bf.Bitfield
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewSingleFileHandler(info *TorInfo) *SingleFileHandler {
	rw := NewReaderWriter(info.Files())

	return &SingleFileHandler{
		info:   info,
		fentry: &info.Files()[0],
		rw:     rw,
		bf:     bf.NewBitfield(info.NumPieces()),
	}
}

// ============================================================================
// IMPL =======================================================================

func (sfh *SingleFileHandler) Piece(index int64, buf []byte) (int64, error) {
	info := sfh.info

	if index >= info.numPieces {
		return 0, fmt.Errorf("attempted to get index %v, max is %v", index, info.numPieces-1)
	}

	seekAmnt := info.pieceLen * index
	readAmnt := info.pieceLen
	truncAmnt := info.length % info.pieceLen

	// Last piece may be truncated
	if index == info.numPieces-1 && truncAmnt != 0 {
		readAmnt = truncAmnt
	}

	if int64(len(buf)) < readAmnt {
		return 0, fmt.Errorf("buffer to small, need %v, got %v", readAmnt, len(buf))
	}

	_, e := sfh.rw.Read(sfh.fentry.LocalPath(), seekAmnt, buf[:readAmnt])
	return readAmnt, e
}

func (sfh *SingleFileHandler) Write(index int64, data []byte) error {
	info := sfh.info
	if index >= info.numPieces {
		return fmt.Errorf("index out of bounds, got %v, max %v", index, info.numPieces)
	}

	hash := utils.SHA1(data)
	knownHash, _ := info.PieceHash(index)

	if hash != knownHash {
		return &HashError{}
	}

	seekAmnt := index * info.pieceLen

	e := sfh.rw.Write(sfh.fentry.LocalPath(), seekAmnt, data)

	return e
}

func (sfh *SingleFileHandler) Validate() error {

	const mib = 1048576
	const maxBuf = 10 * mib

	info := sfh.TorInfo()
	npieces := maxBuf / info.PieceLen()
	bufSize := npieces * info.PieceLen()

	buf := make([]byte, bufSize, bufSize)
	seek := int64(0)
	index := int64(0) // Piece index

	readPath := sfh.fentry.LocalPath()

	for {
		n, e := sfh.rw.Read(readPath, seek, buf)
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		seek += n

		pieces := utils.SegmentData(buf, info.PieceLen())
		for _, piece := range pieces {
			wantHash, _ := info.PieceHash(index)
			gotHash := utils.SHA1(piece)
			eq := wantHash == gotHash
			sfh.Bitfield().Set(index, eq)
			index++
		}
	}

	return nil
}

func (sfh *SingleFileHandler) TorInfo() *TorInfo {
	return sfh.info
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
