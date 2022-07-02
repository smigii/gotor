package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func WriteTestFile(fpath string, data []byte) error {
	e := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
	if e != nil {
		return e
	}

	f, e := os.OpenFile(fpath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if e != nil {
		return e
	}

	_, e = f.Write(data)
	if e != nil {
		return e
	}

	e = f.Close()
	if e != nil {
		return e
	}

	return nil
}

// CloseAndClean closes file pointer fp then calls CleanUpTestFile
func CloseAndClean(fp *os.File, fpath string) error {
	e := fp.Close()
	if e != nil {
		return e
	}

	return CleanUpTestFile(fpath)
}

// CleanUpTestFile will call os.RemoveAll on the base directory specified
// in fpath.
func CleanUpTestFile(fpath string) error {
	parts := strings.Split(fpath, "/")
	if len(parts) == 0 {
		return fmt.Errorf("len(parts) == 0 for input [%v]", fpath)
	}

	e := os.RemoveAll(parts[0]) // Clean up
	if e != nil {
		return fmt.Errorf("RemoveAll(%v) failed on input [%v] - %v", parts[0], fpath, e)
	}

	return nil
}

var dummyHash []byte = nil

// DummyHashes creates and returns a string of length n*20, to be used for
// creating dummy hash strings needed for testing.
func DummyHashes(n int64) string {
	size := 20 * n
	if dummyHash == nil || int64(len(dummyHash)) < size {
		dummyHash = make([]byte, size, size)
	}

	return string(dummyHash[:size])
}

func CheckError(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func CheckFatal(t *testing.T, e error) {
	if e != nil {
		t.Fatal(e)
	}
}
