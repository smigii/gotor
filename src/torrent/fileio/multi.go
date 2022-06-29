package fileio

import (
	"gotor/bf"
	"gotor/utils"
)

// ============================================================================
// STRUCTS ====================================================================

type MultiFileHandler struct {
	files []FileEntryWrapper
	rw    *readerWriter
	meta  *TorFileMeta
	bf    *bf.Bitfield
}

// ============================================================================
// GETTERS ====================================================================

func (mfh *MultiFileHandler) Files() []FileEntryWrapper {
	return mfh.files
}

func (mfh *MultiFileHandler) FileMeta() *TorFileMeta {
	return mfh.meta
}

func (mfh *MultiFileHandler) Bitfield() *bf.Bitfield {
	return mfh.bf
}

// ============================================================================
// FUNC =======================================================================

func (mfh *MultiFileHandler) Piece(index int64) ([]byte, error) {
	files := mfh.GetFiles(index)
	piece := make([]byte, mfh.meta.pieceLen, mfh.meta.pieceLen)
	off := int64(0)

	for _, fe := range files {
		pInfo, e := fe.PieceInfo(index, mfh.meta.pieceLen)
		if e != nil {
			return nil, e
		}

		subpiece := piece[off : off+pInfo.ReadAmnt]
		e = mfh.rw.Read(fe.fpath, pInfo.SeekAmnt, subpiece)
		if e != nil {
			return nil, e
		}

		off += pInfo.ReadAmnt
	}

	return piece[:off], nil
}

func (mfh *MultiFileHandler) Write(index int64, data []byte) error {

	wantHash, e := mfh.meta.PieceHash(index)
	if e != nil {
		return e
	}

	if utils.SHA1(data) != wantHash {
		return &HashError{}
	}

	off := int64(0)
	files := mfh.GetFiles(index)
	for _, fe := range files {
		pInfo, e := fe.PieceInfo(index, mfh.meta.pieceLen)
		if e != nil {
			return e
		}

		subpiece := data[off : off+pInfo.ReadAmnt]
		e = mfh.rw.Write(fe.Path(), pInfo.SeekAmnt, subpiece)
		if e != nil {
			return e
		}

		off += pInfo.ReadAmnt
	}

	return nil
}

func (mfh *MultiFileHandler) Validate() error {

	return nil
}

func (mfh *MultiFileHandler) Close() error {
	return mfh.rw.CloseAll()
}

func NewMultiFileHandler(meta *TorFileMeta) (*MultiFileHandler, error) {
	rw, e := NewReaderWriter(meta.files)
	if e != nil {
		return nil, e
	}

	flist := MultiFileHandler{
		files: make([]FileEntryWrapper, 0, len(meta.files)),
		meta:  meta,
		bf:    bf.NewBitfield(meta.numPieces),
		rw:    rw,
	}

	// This should be 0 by default, since multi-file torrents
	// shouldn't have a "length" key in their info dictionary
	meta.length = 0

	index := int64(0)  // Piece index
	offset := int64(0) // Offset within index

	for _, tfe := range meta.files {

		startPiece := index
		startOff := offset
		endPiece := index + ((tfe.length - 1) / meta.pieceLen)
		endOff := offset + ((tfe.length - 1) % meta.pieceLen)
		if endOff >= meta.pieceLen {
			endPiece += endOff / meta.pieceLen
			endOff %= meta.pieceLen
		}

		index = endPiece
		offset = endOff + 1
		if offset == meta.pieceLen {
			index++
			offset = 0
		}

		flist.meta.length += tfe.length

		flist.files = append(flist.files, FileEntryWrapper{
			FileEntry:     tfe,
			startPieceIdx: startPiece,
			endPieceIdx:   endPiece,
			startPieceOff: startOff,
			endPieceOff:   endOff,
		})
	}

	return &flist, nil
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
