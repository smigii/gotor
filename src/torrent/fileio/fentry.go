package fileio

import (
	"fmt"
	"io"
	"os"
)

// ============================================================================
// STRUCTS ====================================================================

// FileEntry holds only the data contained in a torrent's info dictionary.
// This exists to facilitate testing; rather than passing in a bencode.List to
// testing setup, we can pass these instead.
type FileEntry struct {
	length int64
	fpath  string
}

// FileEntryWrapper wraps the FileEntry struct with the starting and ending
// indices and byte-offsets for a file. This is used by MultiFileHandler structs to
// find which files are related to which pieces.
type FileEntryWrapper struct {
	FileEntry
	startPieceIdx int64 // Starting piece index
	endPieceIdx   int64 // Last piece index (inclusive)
	startPieceOff int64 // Offset from start of startPieceIdx
	endPieceOff   int64 // Offset from start of endPieceIdx (inclusive)
}

// ============================================================================
// GETTERS ====================================================================

func (fe *FileEntry) Length() int64 {
	return fe.length
}

func (fe *FileEntry) Path() string {
	return fe.fpath
}

// ----------------------------------------------------------------------------

func (few *FileEntryWrapper) StartPiece() int64 {
	return few.startPieceIdx
}

func (few *FileEntryWrapper) EndPiece() int64 {
	return few.endPieceIdx
}

func (few *FileEntryWrapper) StartPieceOff() int64 {
	return few.startPieceOff
}

func (few *FileEntryWrapper) EndPieceOff() int64 {
	return few.endPieceOff
}

// ============================================================================
// FUNK =======================================================================

// GetPiece writes the file data of the specified piece index to the dst byte
// slice.
func (few *FileEntryWrapper) GetPiece(dst []byte, index int64, pieceLen int64) (int64, error) {
	/* ===================================================================

	Consider the following,

	4 pieces, each of length 3, which make up 3 files (A, B, C)
	[A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
	Pieces are 0-indexed.

	----------------------------------------------------------------------

	Say we call GetPiece() on file B for piece 2

	File B starts at piece 1. We need to skip the first two bytes
	in the file so that we are pointing to piece 2, then write those
	3 bytes to dst.

	----------------------------------------------------------------------

	Now say we call GetPiece on file B for piece 3

	After skipping over 5 bytes (2 of piece 1, 3 of piece 2), we need to
	make sure we only write a single byte, as the remainder of the piece
	is in file C.

	=================================================================== */

	if index < few.startPieceIdx || index > few.endPieceIdx {
		return 0, fmt.Errorf("index %v out of range", index)
	}

	seekAmnt := int64(0)
	readAmnt := int64(0)

	cursorIdx := few.startPieceIdx
	cursorOff := few.startPieceOff

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
	if index < few.endPieceIdx {
		// We can just write pieceLen bytes, as the file contains more pieces.
		readAmnt = pieceLen
	} else {
		// The remainder of the piece is in a later file.
		readAmnt = few.endPieceOff + 1
	}

	if int64(len(dst)) < readAmnt {
		readAmnt = int64(len(dst))
	}

	f, e := os.Open(few.fpath)
	if e != nil {
		return 0, e
	}

	_, e = f.Seek(seekAmnt, io.SeekStart)
	if e != nil {
		return 0, e
	}

	n, e := f.Read(dst[:readAmnt])
	if e != nil {
		return 0, e
	}
	if int64(n) != readAmnt {
		return int64(n), fmt.Errorf("only read %v bytes, should have read %v", n, readAmnt)
	}

	e = f.Close()
	if e != nil {
		return 0, e
	}

	return readAmnt, nil
}
