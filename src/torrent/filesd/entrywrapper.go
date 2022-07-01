package filesd

import (
	"fmt"
)

// ============================================================================
// STRUCTS ====================================================================

// EntryWrapper wraps the Entry struct with the starting and ending
// indices and byte-offsets for a file. This is used by MultiFileHandler
// structs to find which files are related to which pieces.
type EntryWrapper struct {
	Entry
	startPieceIdx int64 // Starting piece index
	endPieceIdx   int64 // Last piece index (inclusive)
	startPieceOff int64 // Offset from start of startPieceIdx
	endPieceOff   int64 // Offset from start of endPieceIdx (inclusive)
}

// PieceInfo holds the SEEK_SET seek offset and amount to read for a given
// piece index. For torrent [A|A|A] [A|B|B] [B|B|B] [B|C|C], PieceInfo(1) on
// file A would give {3, 1}, as we need to seek 3 bytes, and only 1 byte of the
// piece belongs to A.
type PieceInfo struct {
	SeekAmnt int64
	ReadAmnt int64
}

// ============================================================================
// GETTERS ====================================================================

func (few *EntryWrapper) StartPiece() int64 {
	return few.startPieceIdx
}

func (few *EntryWrapper) EndPiece() int64 {
	return few.endPieceIdx
}

func (few *EntryWrapper) StartPieceOff() int64 {
	return few.startPieceOff
}

func (few *EntryWrapper) EndPieceOff() int64 {
	return few.endPieceOff
}

// ============================================================================
// FUNK =======================================================================

func MakeEntryWrapper(en Entry, startIdx, endIdx, startOff, endOff int64) EntryWrapper {
	return EntryWrapper{
		Entry:         en,
		startPieceIdx: startIdx,
		endPieceIdx:   endIdx,
		startPieceOff: startOff,
		endPieceOff:   endOff,
	}
}

// PieceInfo calculates and returns PieceInfo struct for a given piece index.
func (few *EntryWrapper) PieceInfo(index int64, pieceLen int64) (*PieceInfo, error) {
	/* ===================================================================

	Consider the following,

	4 pieces, each of length 3, which make up 3 files (A, B, C)
	[A|A|A]  [A|B|B]  [B|B|B]  [B|C|C]
	Pieces are 0-indexed.

	----------------------------------------------------------------------

	Say we call PieceInfo() on file B for piece 2

	File B starts at piece 1. We need to skip the first two bytes
	in the file so that we are pointing to piece 2, then write those
	3 bytes to dst.

	----------------------------------------------------------------------

	Now say we call PieceInfo on file B for piece 3

	After skipping over 5 bytes (2 of piece 1, 3 of piece 2), we need to
	make sure we only write a single byte, as the remainder of the piece
	is in file C.

	=================================================================== */

	if index < few.startPieceIdx || index > few.endPieceIdx {
		return nil, fmt.Errorf("index %v out of range", index)
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
		// The file contains greater piece indices
		readAmnt = pieceLen - cursorOff
	} else {
		// The file contains no greater piece indices
		readAmnt = (few.endPieceOff - cursorOff) + 1
	}

	info := PieceInfo{
		SeekAmnt: seekAmnt,
		ReadAmnt: readAmnt,
	}

	return &info, nil
}
