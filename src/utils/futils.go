package utils

import (
	"io"
	"os"
	"path/filepath"
)

// writeSize is the number of bytes written to a file at a time
const writeSize = 1048576

// CreateZeroFilledFile writes an empty file of specified size.
func CreateZeroFilledFile(fpath string, size int64) (*os.File, error) {

	e := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
	if e != nil {
		return nil, e
	}

	f, e := os.OpenFile(fpath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if e != nil {
		return nil, e
	}

	e = AppendZeros(f, size)
	if e != nil {
		return nil, e
	}

	return f, nil
}

// AppendZeros appends amnt zero bytes (0x0) to file fp.
func AppendZeros(fp *os.File, amnt int64) error {
	var e error
	left := amnt
	data := make([]byte, writeSize) // 1MiB write

	_, e = fp.Seek(0, io.SeekEnd)
	if e != nil {
		return e
	}

	for {
		if left == 0 {
			break
		}
		if writeSize < left {
			_, e = fp.Write(data)
			left -= writeSize
		} else {
			_, e = fp.Write(data[:left])
			left = 0
		}
		if e != nil {
			return e
		}
	}
	return nil
}

// OpenCheck will try to first open the file and check the size. If the file
// does not exist, it will create the file at the appropriate size. If the
// file is too big, it will truncate it to the correct size. If it is too
// small, it will grow the file.
func OpenCheck(fpath string, length int64) (*os.File, error) {

	fp, e := os.OpenFile(fpath, os.O_RDWR, 0666)

	// Create new file
	if e != nil {
		fp, e = CreateZeroFilledFile(fpath, length)
		if e != nil {
			return nil, e
		}
	}

	// Check size
	s, e := fp.Stat()
	if e != nil {
		return nil, e
	}
	chkLen := s.Size()
	if chkLen == length {
		return fp, nil
	} else if chkLen > length {
		e = fp.Truncate(length)
		if e != nil {
			return nil, e
		} else {
			return fp, nil
		}
	} else {
		e = AppendZeros(fp, length-chkLen)
		if e != nil {
			return nil, e
		}
	}

	return fp, nil
}
