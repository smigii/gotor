package fileio

import "gotor/utils"

// ============================================================================
// STRUCTS ====================================================================

type FileList struct {
	files []FileEntryWrapper
	fmeta *TorFileMeta
	bf    *utils.Bitfield
}

// ============================================================================
// GETTERS ====================================================================

func (fl FileList) Files() []FileEntryWrapper {
	return fl.files
}

func (fl FileList) FileMeta() *TorFileMeta {
	return fl.fmeta
}

func (fl *FileList) Bitfield() *utils.Bitfield {
	return fl.bf
}

// ============================================================================
// FUNC =======================================================================

func (fl *FileList) Piece(index int64) ([]byte, error) {
	files := fl.GetFiles(index)
	piece := make([]byte, fl.fmeta.pieceLen, fl.fmeta.pieceLen)
	off := int64(0)

	for _, fe := range files {
		n, e := fe.GetPiece(piece[off:], index, fl.fmeta.pieceLen)
		if e != nil {
			// TODO: This should be handled better
			return nil, e
		}
		off += n
	}

	return piece[:off], nil
}

func (fl *FileList) Write(index int64, data []byte) error {

	return nil
}

func (fl *FileList) Validate() error {

	return nil
}

func (fl *FileList) Close() error {
	return nil
}

func NewFileList(fmeta *TorFileMeta) *FileList {
	flist := FileList{
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
func (fl *FileList) GetFiles(piece int64) []FileEntryWrapper {

	hit := false
	startIdx := 0
	n := 0

	for i, fe := range fl.Files() {
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

	return fl.files[startIdx : startIdx+n]
}
