package torrent

import (
	"fmt"
	"io"
	"math"
	"os"
)

// ============================================================================
// STRUCTS ====================================================================

type FileEntry struct {
	torFileEntry
	startPieceIdx uint64 // Starting piece index
	endPieceIdx   uint64 // Last piece index (inclusive)
	startPieceOff uint64 // Offset from start of startPieceIdx
	endPieceOff   uint64 // Offset from start of endPieceIdx (inclusive)
}

// ============================================================================
// GETTERS ====================================================================

func (fe FileEntry) Length() uint64 {
	return fe.length
}

func (fe FileEntry) Path() string {
	return fe.fpath
}

func (fe FileEntry) StartPiece() uint64 {
	return fe.startPieceIdx
}

func (fe FileEntry) EndPiece() uint64 {
	return fe.endPieceIdx
}

func (fe FileEntry) StartPieceOff() uint64 {
	return fe.startPieceOff
}

func (fe FileEntry) EndPieceOff() uint64 {
	return fe.endPieceOff
}

// ============================================================================
// FUNK =======================================================================

// GetPiece writes the file data of the specified piece index to the dst byte
// slice.
func (fe *FileEntry) GetPiece(dst []byte, index uint64, pieceLen uint64) (uint64, error) {
	/*

		Consider the following,

		4 pieces, each of length 3, which make up 3 files (A, B, C)
		[A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
		Pieces are 0-indexed.

		-----------------------------------------------------------------------

		Say we call GetPiece() on file B for piece 2

		File B starts at piece 1. We need to skip the first two bytes
		in the file so that we are pointing to piece 2, then write those
		3 bytes to dst.

		-----------------------------------------------------------------------

		Now say we call GetPiece on file B for piece 3

		After skipping over 5 bytes (2 of piece 1, 3 of piece 2), we need to
		make sure we only write a single byte, as the remainder of the piece
		is in file C.

	*/

	if index < fe.startPieceIdx || index > fe.endPieceIdx {
		return 0, fmt.Errorf("index %v out of range", index)
	}

	seekAmnt := uint64(0)
	readAmnt := uint64(0)

	cursorIdx := fe.startPieceIdx
	cursorOff := fe.startPieceOff

	// Adjust seekAmnt
	for {
		if cursorIdx == index {
			break
		}

		seekAmnt += pieceLen - cursorOff
		cursorOff = 0
		cursorIdx++
	}

	// Adjust readAmnt
	if index < fe.endPieceIdx {
		// We can just write pieceLen bytes, as the file contains more pieces.
		readAmnt = pieceLen
	} else {
		// The remainder of the piece is in a later file.
		readAmnt = fe.endPieceOff + 1
	}

	if uint64(len(dst)) < readAmnt {
		readAmnt = uint64(len(dst))
	}

	f, e := os.Open(fe.fpath)
	if e != nil {
		return 0, e
	}

	// Probably overkill
	const maxSeek = uint64(math.MaxInt64)
	for {
		if seekAmnt == 0 {
			break
		} else if seekAmnt > maxSeek {
			_, e = f.Seek(math.MaxInt64, io.SeekCurrent)
			seekAmnt -= maxSeek
		} else {
			_, e = f.Seek(int64(seekAmnt), io.SeekCurrent)
			seekAmnt -= seekAmnt
		}
		if e != nil {
			return 0, e
		}
	}

	n, e := f.Read(dst[:readAmnt])
	if e != nil {
		return 0, e
	}
	if uint64(n) != readAmnt {
		return uint64(n), fmt.Errorf("only read %v bytes, should have read %v", n, readAmnt)
	}

	e = f.Close()
	if e != nil {
		return 0, e
	}

	return readAmnt, nil
}
