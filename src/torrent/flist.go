package torrent

// ============================================================================
// STRUCTS ====================================================================

type FileList struct {
	files    []FileEntry
	piecelen uint64
	length   uint64 // Total length of all files
}

// torFileEntry holds only the data contained in a torrent file. This exists
// so that we can separate extracting the data from the torrent file and
// calculating the index/offset data.
type torFileEntry struct {
	length uint64
	fpath  string
}

// ============================================================================
// GETTERS ====================================================================

func (fl FileList) Files() []FileEntry {
	return fl.files
}

func (fl FileList) Length() uint64 {
	return fl.length
}

// ============================================================================
// FUNC =======================================================================

func newFileList(torFileEntries []torFileEntry, piecelen uint64) *FileList {
	flist := FileList{
		files:    make([]FileEntry, 0, len(torFileEntries)),
		piecelen: piecelen,
		length:   0,
	}

	index := uint64(0)  // Piece index
	offset := uint64(0) // Offset within index

	for _, tfe := range torFileEntries {

		startPiece := index
		startOff := offset
		endPiece := index + ((tfe.length - 1) / piecelen)
		endOff := offset + ((tfe.length - 1) % piecelen)
		if endOff >= piecelen {
			endPiece += endOff / piecelen
			endOff %= piecelen
		}

		index = endPiece
		offset = endOff + 1
		if offset == piecelen {
			index++
			offset = 0
		}

		flist.length += tfe.length

		flist.files = append(flist.files, FileEntry{
			torFileEntry:  tfe,
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
func (fl *FileList) GetFiles(piece uint64) []FileEntry {

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

// Piece returns the file data of given piece index.
func (fl *FileList) Piece(index uint64) ([]byte, error) {
	files := fl.GetFiles(index)
	piece := make([]byte, fl.piecelen, fl.piecelen)
	off := uint64(0)

	for _, fe := range files {
		n, e := fe.GetPiece(piece[off:], index, fl.piecelen)
		if e != nil {
			// TODO: This should be handled better
			return nil, e
		}
		off += n
	}

	return piece[:off], nil
}

// Write writes the given data to the file(s) corresponding to piece index.
func (fl *FileList) Write(index uint64, data []byte) error {

	return nil
}
