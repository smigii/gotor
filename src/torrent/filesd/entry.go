package filesd

import (
	"strings"

	"gotor/bencode"
)

// ============================================================================
// STRUCTS ====================================================================

// Entry represents an entry in a torrent's info["files"] list, which holds
// dictionaries of {length, path} for each file. Entry also includes the
// localPath field, which is can be changed by the user to change the location
// or name of the file.
type Entry struct {
	length    int64
	torPath   string // File path as defined in torrent file
	localPath string // File path as defined by user (optional)
}

// ============================================================================
// GETTERS ====================================================================

func (fe *Entry) Length() int64 {
	return fe.length
}

func (fe *Entry) TorPath() string {
	return fe.torPath
}

func (fe *Entry) LocalPath() string {
	return fe.localPath
}

func (fe *Entry) SetLocalPath(newPath string) {
	fe.localPath = newPath
}

// ============================================================================
// FUNK =======================================================================

func MakeFileEntry(torPath string, length int64) Entry {
	return Entry{
		length:    length,
		torPath:   torPath,
		localPath: torPath,
	}
}

func (fe *Entry) Bencode() bencode.Dict {
	pathTokens := strings.Split(fe.TorPath(), "/")
	pathList := make(bencode.List, 0, len(pathTokens))

	for _, pathToken := range pathTokens {
		pathList = append(pathList, pathToken)
	}

	d := make(bencode.Dict)
	d["length"] = fe.Length()
	d["path"] = pathList

	return d
}

func CalcNumPieces(files []Entry, pieceLen int64) int64 {
	l := int64(0)

	for _, fe := range files {
		l += fe.Length()
	}

	n := l / pieceLen
	if l%pieceLen > 0 {
		n++
	}

	return n
}
