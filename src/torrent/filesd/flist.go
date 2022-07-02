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

// GetFiles returns all files that are contained within the specified piece
// index.
func (fl FileList) GetFiles(piece int64) []Entry {

	hit := false
	startIdx := 0
	n := 0

	for i, fe := range fl {
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

	return fl[startIdx : startIdx+n]
}
