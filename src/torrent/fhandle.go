package torrent

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// writeSize is the number of bytes written to a file at a time
const writeSize = 1048576

type FileHandler struct {
	wd       string // Working directory
	filelist *FileList
}

func NewFileHandler(wd string, flist *FileList) *FileHandler {
	return &FileHandler{
		wd:       wd,
		filelist: flist,
	}
}

// Piece returns the file data of given piece index.
func (fh *FileHandler) Piece(index uint64) ([]byte, error) {
	//files := fh.filelist.GetFiles(index)
	//strb := strings.Builder{}

	//for _, fe := range files {
	//	fe.startPieceIdx
	//}

	return nil, nil
}

// Write writes the given data to the file(s) corresponding to piece index.
func (fh *FileHandler) Write(index uint64, data []byte) error {

	return nil
}

// Validate will look through all the files specified in the torrent and check
// the pieces and their hashes. If a file doesn't exist, the file will be
// created and set to the correct size. If a file exists, but is the wrong
// size, empty bytes will be appended to the correct size. If all the files
// are correct, returns true, else returns false.
func (fh *FileHandler) Validate() bool {

	valid := true

	for _, fe := range fh.filelist.Files() {
		fmt.Println(path.Join(fh.wd, fe.fpath))
	}

	return valid
}

// writeEmptyFile writes an empty file of specified size.
func writeEmptyFile(fpath string, size uint64) error {

	e := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
	if e != nil {
		return e
	}

	f, e := os.Create(fpath)
	if e != nil {
		return e
	}

	left := size
	data := make([]byte, writeSize) // 1MiB write
	for {
		if left == 0 {
			break
		}
		if writeSize < left {
			_, e = f.Write(data)
			left -= writeSize
		} else {
			_, e = f.Write(data[:left])
			left = 0
		}
		if e != nil {
			return e
		}
	}

	return nil
}
