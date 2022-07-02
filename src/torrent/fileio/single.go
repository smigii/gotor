package fileio

import (
	"fmt"
	"io"

	"gotor/bf"
	fentry2 "gotor/torrent/filesd"
	"gotor/torrent/info"
	"gotor/utils"
)

// ============================================================================
// STRUCT =====================================================================

// SingleFileHandler is used for single-file torrents. It can access pieces
// faster than MultiFileHandler since it doesn't need to find out which files
// are contained in a given piece.
type SingleFileHandler struct {
	tinfo  *info.TorInfo
	fentry *fentry2.EntryBase
	rw     *readerWriter
	bf     *bf.Bitfield
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewSingleFileHandler(tinfo *info.TorInfo) *SingleFileHandler {
	rw := NewReaderWriter(tinfo.Files())

	return &SingleFileHandler{
		tinfo:  tinfo,
		fentry: &tinfo.Files()[0],
		rw:     rw,
		bf:     bf.NewBitfield(tinfo.NumPieces()),
	}
}

// ============================================================================
// IMPL =======================================================================

func (sfh *SingleFileHandler) Piece(index int64, buf []byte) (int64, error) {
	tinfo := sfh.tinfo

	if index >= tinfo.NumPieces() {
		return 0, fmt.Errorf("attempted to get index %v, max is %v", index, tinfo.NumPieces()-1)
	}

	seekAmnt := tinfo.PieceLen() * index
	readAmnt := tinfo.PieceLen()
	truncAmnt := tinfo.Length() % tinfo.PieceLen()

	// Last piece may be truncated
	if index == tinfo.NumPieces()-1 && truncAmnt != 0 {
		readAmnt = truncAmnt
	}

	if int64(len(buf)) < readAmnt {
		return 0, fmt.Errorf("buffer to small, need %v, got %v", readAmnt, len(buf))
	}

	_, e := sfh.rw.Read(sfh.fentry.LocalPath(), seekAmnt, buf[:readAmnt])
	return readAmnt, e
}

func (sfh *SingleFileHandler) Write(index int64, data []byte) error {
	tinfo := sfh.tinfo
	if index >= tinfo.NumPieces() {
		return fmt.Errorf("index out of bounds, got %v, max %v", index, tinfo.NumPieces())
	}

	hash := utils.SHA1(data)
	knownHash, _ := tinfo.PieceHash(index)

	if hash != knownHash {
		return &HashError{}
	}

	seekAmnt := index * tinfo.PieceLen()

	e := sfh.rw.Write(sfh.fentry.LocalPath(), seekAmnt, data)

	return e
}

func (sfh *SingleFileHandler) Validate() error {

	const mib = 1048576
	const maxBuf = 10 * mib

	tinfo := sfh.TorInfo()
	npieces := maxBuf / tinfo.PieceLen()
	bufSize := npieces * tinfo.PieceLen()

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

		pieces := utils.SegmentData(buf, tinfo.PieceLen())
		for _, piece := range pieces {
			wantHash, _ := tinfo.PieceHash(index)
			gotHash := utils.SHA1(piece)
			eq := wantHash == gotHash
			sfh.Bitfield().Set(index, eq)
			index++
		}
	}

	return nil
}

func (sfh *SingleFileHandler) TorInfo() *info.TorInfo {
	return sfh.tinfo
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
