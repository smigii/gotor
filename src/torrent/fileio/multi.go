package fileio

import (
	"gotor/bf"
	"gotor/torrent/filesd"
	"gotor/torrent/info"
	"gotor/utils"
)

// ============================================================================
// STRUCTS ====================================================================

type MultiFileHandler struct {
	files []filesd.Entry
	tinfo *info.TorInfo
	rw    *readerWriter
	bf    *bf.Bitfield
}

// ============================================================================
// GETTERS ====================================================================

func (mfh *MultiFileHandler) Files() []filesd.Entry {
	return mfh.files
}

func (mfh *MultiFileHandler) TorInfo() *info.TorInfo {
	return mfh.tinfo
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
		pInfo, e := fe.PieceInfo(index, mfh.tinfo.PieceLen())
		if e != nil {
			return 0, e
		}

		if int64(len(buf)) < (off + pInfo.ReadAmnt) {
			panic("buffer not long enough")
		}

		subpiece := buf[off : off+pInfo.ReadAmnt]
		_, e = mfh.rw.Read(fe.LocalPath(), pInfo.SeekAmnt, subpiece)
		if e != nil {
			return 0, e
		}

		off += pInfo.ReadAmnt
	}

	return off, nil
}

func (mfh *MultiFileHandler) Write(index int64, data []byte) error {

	wantHash, e := mfh.tinfo.PieceHash(index)
	if e != nil {
		return e
	}

	if utils.SHA1(data) != wantHash {
		return &HashError{}
	}

	off := int64(0)
	files := mfh.GetFiles(index)
	for _, fe := range files {
		pInfo, e := fe.PieceInfo(index, mfh.tinfo.PieceLen())
		if e != nil {
			return e
		}

		subpiece := data[off : off+pInfo.ReadAmnt]
		e = mfh.rw.Write(fe.LocalPath(), pInfo.SeekAmnt, subpiece)
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

func NewMultiFileHandler(tinfo *info.TorInfo) *MultiFileHandler {
	rw := NewReaderWriter(tinfo.Files())

	mfh := MultiFileHandler{
		files: make([]filesd.Entry, 0, len(tinfo.Files())),
		tinfo: tinfo,
		bf:    bf.NewBitfield(tinfo.NumPieces()),
		rw:    rw,
	}

	index := int64(0)  // Piece index
	offset := int64(0) // Offset within index

	for _, tfe := range tinfo.Files() {

		startPiece := index
		startOff := offset
		endPiece := index + ((tfe.Length() - 1) / tinfo.PieceLen())
		endOff := offset + ((tfe.Length() - 1) % tinfo.PieceLen())
		if endOff >= tinfo.PieceLen() {
			endPiece += endOff / tinfo.PieceLen()
			endOff %= tinfo.PieceLen()
		}

		index = endPiece
		offset = endOff + 1
		if offset == tinfo.PieceLen() {
			index++
			offset = 0
		}

		ew := filesd.MakeEntryWrapper(tfe, startPiece, endPiece, startOff, endOff)
		mfh.files = append(mfh.files, ew)
	}

	return &mfh
}

// GetFiles returns all files that are contained within the specified piece
// index.
func (mfh *MultiFileHandler) GetFiles(piece int64) []filesd.Entry {

	hit := false
	startIdx := 0
	n := 0

	for i, fe := range mfh.Files() {
		if fe.StartPiece() > piece {
			break
		}

		if fe.StartPiece() <= piece && fe.EndPiece() >= piece {
			if !hit {
				startIdx = i
				hit = true
			}
			n += 1
		}
	}

	return mfh.files[startIdx : startIdx+n]
}
