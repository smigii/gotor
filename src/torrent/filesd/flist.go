package filesd

type FileList []Entry

func MakeFileList(entries []EntryBase, pieceLen int64) FileList {

	fl := make(FileList, 0, 2)

	index := int64(0)  // Piece index
	offset := int64(0) // Offset within index

	for _, tfe := range entries {

		startPiece := index
		startOff := offset
		endPiece := index + ((tfe.Length() - 1) / pieceLen)
		endOff := offset + ((tfe.Length() - 1) % pieceLen)
		if endOff >= pieceLen {
			endPiece += endOff / pieceLen
			endOff %= pieceLen
		}

		index = endPiece
		offset = endOff + 1
		if offset == pieceLen {
			index++
			offset = 0
		}

		ew := MakeEntryWrapper(tfe, startPiece, endPiece, startOff, endOff)
		fl = append(fl, ew)
	}

	return fl
}
