package utils

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func WriteTestFile(fpath string, data []byte) error {
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
