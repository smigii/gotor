package filesd

import (
	"errors"
	"fmt"
	"strings"

	"gotor/bencode"
)

// ============================================================================
// STRUCTS ====================================================================

// EntryBase represents an entry in a torrent's info["files"] list, which holds
// dictionaries of {length, path} for each file. EntryBase also includes the
// localPath field, which is can be changed by the user to change the location
// or name of the file.
type EntryBase struct {
	length    int64
	torPath   string // File path as defined in torrent file
	localPath string // File path as defined by user (optional)
}

// ============================================================================
// GETTERS ====================================================================

func (fe *EntryBase) Length() int64 {
	return fe.length
}

func (fe *EntryBase) TorPath() string {
	return fe.torPath
}

func (fe *EntryBase) LocalPath() string {
	return fe.localPath
}

func (fe *EntryBase) SetLocalPath(newPath string) {
	fe.localPath = newPath
}

// ============================================================================
// FUNK =======================================================================

func MakeFileEntry(torPath string, length int64) EntryBase {
	return EntryBase{
		length:    length,
		torPath:   torPath,
		localPath: torPath,
	}
}

func FromBenList(benlist bencode.List) ([]EntryBase, error) {
	entries := make([]EntryBase, 0, 1)

	for _, fEntry := range benlist {
		fDict, ok := fEntry.(bencode.Dict)
		if !ok {
			return nil, errors.New("failed to convert file entry to dictionary")
		}

		fLen, err := fDict.GetInt("length")
		if err != nil {
			return nil, err
		}

		fPathList, err := fDict.GetList("path")
		if err != nil {
			return nil, err
		}

		// Read through list of path strings
		strb := strings.Builder{}

		for _, fPathEntry := range fPathList {
			pathPiece, ok := fPathEntry.(string)
			if !ok {
				return nil, fmt.Errorf(fmt.Sprintf("file entry contains invalid path [%v]", fPathEntry))
			}
			strb.WriteString(pathPiece)
			strb.WriteByte('/')
		}
		l := len(strb.String())

		// exclude last '/'
		entries = append(entries, MakeFileEntry(strb.String()[:l-1], fLen))
	}

	return entries, nil
}

func (fe *EntryBase) Bencode() bencode.Dict {
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

func CalcNumPieces(files []EntryBase, pieceLen int64) int64 {
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
