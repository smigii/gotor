package torrent

// ============================================================================
// STRUCTS ====================================================================

type FileList struct {
	files  []FileEntry
	length uint64 // Total length of all files
}

type FileEntry struct {
	torFileEntry
	startPieceIdx uint64 // Starting piece index
	endPieceIdx   uint64 // Last piece index (inclusive)
	startPieceOff uint64 // Offset from start of startPieceIdx
	endPieceOff   uint64 // Offset from start of endPieceIdx (inclusive)
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

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

func (f FileEntry) Length() uint64 {
	return f.length
}

func (f FileEntry) Path() string {
	return f.fpath
}

func (f FileEntry) StartPiece() uint64 {
	return f.startPieceIdx
}

func (f FileEntry) EndPiece() uint64 {
	return f.endPieceIdx
}

func (f FileEntry) StartPieceOff() uint64 {
	return f.startPieceOff
}

func (f FileEntry) EndPieceOff() uint64 {
	return f.endPieceOff
}

// ============================================================================
// FUNC =======================================================================

func newFileList(torFileEntries []torFileEntry, piecelen uint64) *FileList {
	flist := FileList{
		files:  make([]FileEntry, 0, len(torFileEntries)),
		length: 0,
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

// GetFiles returns all files that are contained with the specified piece
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
