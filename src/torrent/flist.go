package torrent

import (
	"fmt"
	"gotor/bencode"
	"strings"
)

// ============================================================================
// STRUCTS ====================================================================

type FileList struct {
	files  []FileEntry
	length uint64 // Total length of all files
}

type FileEntry struct {
	length        uint64
	path          string
	startPieceIdx uint64 // Starting piece index
	endPieceIdx   uint64 // Last piece index (inclusive)
	startPieceOff uint64 // Offset from start of startPieceIdx
	endPieceOff   uint64 // Offset from start of endPieceIdx (inclusive)
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
	return f.path
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

func newFileList(benlist bencode.List, tor *Torrent) (*FileList, error) {
	flist := FileList{
		files:  make([]FileEntry, 0, len(benlist)),
		length: 0,
	}

	index := uint64(0)  // Piece index
	offset := uint64(0) // Offset within index

	for _, fEntry := range benlist {
		fDict, ok := fEntry.(bencode.Dict)
		if !ok {
			return nil, &TorError{
				msg: fmt.Sprintf("failed to convert file entry to dictionary\n%v", fEntry),
			}
		}

		fLen, err := fDict.GetUint("length")
		if err != nil {
			return nil, err
		}
		flist.length += fLen

		fPathList, err := fDict.GetList("path")
		if err != nil {
			return nil, err
		}

		startPiece := index
		startOff := offset
		endPiece := index + ((fLen - 1) / tor.pieceLen)
		endOff := offset + ((fLen - 1) % tor.pieceLen)
		if endOff >= tor.pieceLen {
			endPiece += endOff / tor.pieceLen
			endOff %= tor.pieceLen
		}

		index = endPiece
		offset = endOff + 1
		if offset == tor.pieceLen {
			index++
			offset = 0
		}

		// Read through list of path strings
		strb := strings.Builder{}

		// Write the directory name
		strb.WriteString(tor.name)
		strb.WriteByte('/')

		for _, fPathEntry := range fPathList {
			pathPiece, ok := fPathEntry.(string)
			if !ok {
				return nil, &TorError{
					msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
				}
			}
			strb.WriteString(pathPiece)
			strb.WriteByte('/')
		}
		l := len(strb.String())

		flist.files = append(flist.files, FileEntry{
			length:        fLen,
			path:          strb.String()[:l-1], // exclude last '/'
			startPieceIdx: startPiece,
			endPieceIdx:   endPiece,
			startPieceOff: startOff,
			endPieceOff:   endOff,
		})
	}

	return &flist, nil
}

// GetFiles returns all files that are contained with the specified piece
// index.
func (fl *FileList) GetFiles(piece uint64) []FileEntry {

	hit := false
	done := false
	startIdx := 0
	endIdx := 0

	for i, fe := range fl.Files() {
		if hit && fe.startPieceIdx > piece {
			endIdx = i
			done = true
			break
		} else if !hit && fe.startPieceIdx == piece {
			startIdx = i
			hit = true
		}
	}

	if !hit {
		return nil
	} else if done {
		return fl.files[startIdx:endIdx]
	} else {
		return fl.files[startIdx:]
	}

}
