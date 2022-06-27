package fileio

import "gotor/utils"

// ============================================================================
// STRUCTS ====================================================================

type MultiFileHandler struct {
	files []FileEntryWrapper
	fmeta *TorFileMeta
	bf    *utils.Bitfield
}

// ============================================================================
// GETTERS ====================================================================

func (mfh *MultiFileHandler) Files() []FileEntryWrapper {
	return mfh.files
}

func (mfh *MultiFileHandler) FileMeta() *TorFileMeta {
	return mfh.fmeta
}

func (mfh *MultiFileHandler) Bitfield() *utils.Bitfield {
	return mfh.bf
}

// ============================================================================
// FUNC =======================================================================

func (mfh *MultiFileHandler) Piece(index int64) ([]byte, error) {
	files := mfh.GetFiles(index)
	piece := make([]byte, mfh.fmeta.pieceLen, mfh.fmeta.pieceLen)
	off := int64(0)

	for _, fe := range files {
		n, e := fe.GetPiece(piece[off:], index, mfh.fmeta.pieceLen)
		if e != nil {
			// TODO: This should be handled better
			return nil, e
		}
		off += n
	}

	return piece[:off], nil
}

func (mfh *MultiFileHandler) Write(index int64, data []byte) error {

	return nil
}

func (mfh *MultiFileHandler) Validate() error {

	return nil
}

func (mfh *MultiFileHandler) Close() error {
	return nil
}

func NewFileList(fmeta *TorFileMeta) *MultiFileHandler {
	flist := MultiFileHandler{
		files: make([]FileEntryWrapper, 0, len(fmeta.files)),
		fmeta: fmeta,
	}

	// This should be 0 by default, since multi-file torrents
	// shouldn't have a "length" key in their info dictionary
	fmeta.length = 0

	index := int64(0)  // Piece index
	offset := int64(0) // Offset within index

	for _, tfe := range fmeta.files {

		startPiece := index
		startOff := offset
		endPiece := index + ((tfe.length - 1) / fmeta.pieceLen)
		endOff := offset + ((tfe.length - 1) % fmeta.pieceLen)
		if endOff >= fmeta.pieceLen {
			endPiece += endOff / fmeta.pieceLen
			endOff %= fmeta.pieceLen
		}

		index = endPiece
		offset = endOff + 1
		if offset == fmeta.pieceLen {
			index++
			offset = 0
		}

		flist.fmeta.length += tfe.length

		flist.files = append(flist.files, FileEntryWrapper{
			FileEntry:     tfe,
			startPieceIdx: startPiece,
			endPieceIdx:   endPiece,
			startPieceOff: startOff,
			endPieceOff:   endOff,
		})
	}

	return &flist
}

// GetFiles returns all files that are contained within the specified piece
// index.
func (mfh *MultiFileHandler) GetFiles(piece int64) []FileEntryWrapper {

	hit := false
	startIdx := 0
	n := 0

	for i, fe := range mfh.Files() {
		if fe.startPieceIdx > piece {
			break
		}

		if fe.startPieceIdx <= piece && fe.endPieceIdx >= piece {
			if !hit {
				startIdx = i
				hit = true
			}
			n += 1
		}
	}

	return mfh.files[startIdx : startIdx+n]
}
