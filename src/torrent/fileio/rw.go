package fileio

import (
	"fmt"
	"os"

	"gotor/torrent/filesd"
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

type readerWriter struct {
	files []filesd.Entry
	ptrs  map[string]*os.File
}

// ============================================================================
// FUNC =======================================================================

func NewReaderWriter(files []filesd.Entry) *readerWriter {
	rw := readerWriter{
		files: files,
		ptrs:  make(map[string]*os.File),
	}

	return &rw
}

func (rw *readerWriter) Write(fpath string, seekAmnt int64, data []byte) error {

	ptr, ok := rw.ptrs[fpath]
	if ok {
		_, e := ptr.WriteAt(data, seekAmnt)
		return e
	} else {
		return &PathError{fpath: fpath}
	}
}

func (rw *readerWriter) Read(fpath string, seekAmnt int64, buf []byte) (int64, error) {
	ptr, ok := rw.ptrs[fpath]
	if ok {
		n, e := ptr.ReadAt(buf, seekAmnt)
		return int64(n), e
	} else {
		return 0, &PathError{fpath: fpath}
	}
}

func (rw *readerWriter) Move(fromPath string, toPath string) error {

	// Acquire outer lock
	// Acquire all inner locks
	// Move files
	// Unlock inner locks
	// Unlock outer locks

	return nil
}

// OCAT will call utils.OCAT to open/create/append/truncate a file
// to the appropriate size.
func (rw *readerWriter) OCAT() error {

	errors := make([]FileErrorEntry, 0, 4)

	for _, fentry := range rw.files {
		f, e := utils.OCAT(fentry.LocalPath(), fentry.Length())
		if e != nil {
			errors = append(errors, FileErrorEntry{
				fpath: fentry.LocalPath(),
				err:   e,
			})
		} else {
			rw.ptrs[fentry.LocalPath()] = f
		}
	}

	var err error
	if len(errors) > 0 {
		err = &OpenError{errors: errors}
	}
	return err
}

func (rw *readerWriter) Close(fpath string) error {
	ptr, ok := rw.ptrs[fpath]
	if ok {
		e := ptr.Close()
		return e
	} else {
		return &PathError{fpath: fpath}
	}
}

func (rw *readerWriter) CloseAll() error {
	for _, fptr := range rw.ptrs {
		e := fptr.Close()
		if e != nil {
			return e
		}
	}

	return nil
}
