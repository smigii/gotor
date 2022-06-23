package torrent

import (
	"fmt"
	"io"
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
}

// ============================================================================
// CONSTRUCTOR ================================================================

func newFileSingle(meta *TorFileMeta) *FileSingle {
	return &FileSingle{
		meta: meta,
	}
}

// ============================================================================
// IMPL =======================================================================

func (f *FileSingle) Piece(index uint64) ([]byte, error) {

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

	fp, e := os.Open(f.meta.name)
	if e != nil {
		panic(e)
	}

	_, e = fp.Seek(int64(seekAmnt), io.SeekStart)
	if e != nil {
		panic(e)
	}

	_, e = fp.Read(buf)
	if e != nil {
		panic(e)
	}

	return buf, nil
}

func (f *FileSingle) Write(index uint64, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (f *FileSingle) FileMeta() *TorFileMeta {
	return f.meta
}

func (f *FileSingle) Validate(bf *utils.Bitfield) {
	//TODO implement me
}
