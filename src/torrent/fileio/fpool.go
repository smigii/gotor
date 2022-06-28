package fileio

import (
	"fmt"
	"os"

	"gotor/utils"
)

// ============================================================================
// ERRORS =====================================================================

type FileErrorEntry struct {
	fpath string
	err   error
}

type OpenError struct {
	errors []FileErrorEntry
}

func (ce *OpenError) Error() string {
	return "error openning files"
}

type PathError struct {
	fpath string
}

func (pe *PathError) Error() string {
	return fmt.Sprintf("path [%v] does not exist", pe.fpath)
}

// ============================================================================
// STRUCTS ====================================================================

type FilePool struct {
	pool map[string]*os.File
}

// ============================================================================
// FUNC =======================================================================

func NewFilePool(files []FileEntry) (*FilePool, error) {
	filePool := FilePool{
		pool: make(map[string]*os.File),
	}

	errors := make([]FileErrorEntry, 0, 4)

	for _, fentry := range files {
		f, e := utils.OpenCheck(fentry.fpath, fentry.length)
		if e != nil {
			errors = append(errors, FileErrorEntry{
				fpath: fentry.fpath,
				err:   e,
			})
		} else {
			filePool.pool[fentry.fpath] = f
		}
	}

	var err error
	if len(errors) > 0 {
		err = &OpenError{errors: errors}
	}
	return &filePool, err
}

func (fp *FilePool) Write(fpath string, seekAmnt int64, data []byte) error {

	ptr, ok := fp.pool[fpath]
	if ok {
		_, e := ptr.WriteAt(data, seekAmnt)
		return e
	} else {
		return &PathError{fpath: fpath}
	}
}

func (fp *FilePool) Read(fpath string, seekAmnt int64, buf []byte) error {
	ptr, ok := fp.pool[fpath]
	if ok {
		_, e := ptr.ReadAt(buf, seekAmnt)
		return e
	} else {
		return &PathError{fpath: fpath}
	}
}

func (fp *FilePool) Move(fromPath string, toPath string) error {

	// Acquire outer lock
	// Acquire all inner locks
	// Move files
	// Unlock inner locks
	// Unlock outer locks

	return nil
}

func (fp *FilePool) Close(fpath string) error {
	ptr, ok := fp.pool[fpath]
	if ok {
		e := ptr.Close()
		return e
	} else {
		return &PathError{fpath: fpath}
	}
}

func (fp *FilePool) CloseAll() error {
	for _, fptr := range fp.pool {
		e := fptr.Close()
		if e != nil {
			return e
		}
	}

	return nil
}
