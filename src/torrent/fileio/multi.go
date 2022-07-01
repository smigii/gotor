package fileio

import (
	"gotor/bf"
	"gotor/utils"
)

// ============================================================================
// STRUCTS ====================================================================

type MultiFileHandler struct {
	files []FileEntryWrapper
	info  *TorInfo
	rw    *readerWriter
	bf    *bf.Bitfield
}

// ============================================================================
// GETTERS ====================================================================

func (mfh *MultiFileHandler) Files() []FileEntryWrapper {
	return mfh.files
}

func (mfh *MultiFileHandler) TorInfo() *TorInfo {
	return mfh.info
}

func (mfh *MultiFileHandler) Bitfield() *bf.Bitfield {
	return mfh.bf
}

// ============================================================================
// FUNC =======================================================================

func (mfh *MultiFileHandler) Piece(index int64, buf []byte) (int64, error) {
	files := mfh.GetFiles(index)
	off := int64(0)

	for _, fe := range files {
		pInfo, e := fe.PieceInfo(index, mfh.info.pieceLen)
		if e != nil {
			return 0, e
		}

		if int64(len(buf)) < (off + pInfo.ReadAmnt) {
			panic("buffer not long enough")
		}

		subpiece := buf[off : off+pInfo.ReadAmnt]
		_, e = mfh.rw.Read(fe.torPath, pInfo.SeekAmnt, subpiece)
		if e != nil {
			return 0, e
		}

		off += pInfo.ReadAmnt
	}

	return off, nil
}

func (mfh *MultiFileHandler) Write(index int64, data []byte) error {

	wantHash, e := mfh.info.PieceHash(index)
	if e != nil {
		return e
	}

	if utils.SHA1(data) != wantHash {
		return &HashError{}
	}

	off := int64(0)
	files := mfh.GetFiles(index)
	for _, fe := range files {
		pInfo, e := fe.PieceInfo(index, mfh.info.pieceLen)
		if e != nil {
			return e
		}

		subpiece := data[off : off+pInfo.ReadAmnt]
		e = mfh.rw.Write(fe.TorPath(), pInfo.SeekAmnt, subpiece)
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

func (mfh *MultiFileHandler) OCAT() error {
	e := mfh.rw.OCAT()
	return e

	//e = mfh.Validate()
	//return e
}

func (mfh *MultiFileHandler) Close() error {
	return mfh.rw.CloseAll()
}

func NewMultiFileHandler(info *TorInfo) *MultiFileHandler {
	rw := NewReaderWriter(info.files)

	flist := MultiFileHandler{
		files: make([]FileEntryWrapper, 0, len(info.files)),
		info:  info,
		bf:    bf.NewBitfield(info.numPieces),
		rw:    rw,
	}

	// This should be 0 by default, since multi-file torrents
	// shouldn't have a "length" key in their info dictionary
	info.length = 0

	index := int64(0)  // Piece index
	offset := int64(0) // Offset within index

	for _, tfe := range info.files {

		startPiece := index
		startOff := offset
		endPiece := index + ((tfe.length - 1) / info.pieceLen)
		endOff := offset + ((tfe.length - 1) % info.pieceLen)
		if endOff >= info.pieceLen {
			endPiece += endOff / info.pieceLen
			endOff %= info.pieceLen
		}

		index = endPiece
		offset = endOff + 1
		if offset == info.pieceLen {
			index++
			offset = 0
		}

		flist.info.length += tfe.length

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
