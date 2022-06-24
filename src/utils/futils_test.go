package utils

import (
	"os"
	"testing"
)

func TestWriteEmptyFile(t *testing.T) {

	tests := []struct {
		name  string
		fpath string
		len   int64
	}{
		{"Single Write", "TestEmptyWrite", writeSize - 100},
		{"Single Write Exact", "TestEmptyWrite", writeSize},
		{"Single File Multi Write", "TestEmptyWrite", writeSize + 100},
		{"Nested File", "TestEmptyWrite/subdir/test", writeSize + 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, e := CreateZeroFilledFile(tt.fpath, tt.len)
			if e != nil {
				CleanUpTestFile(tt.fpath)
				t.Error(e)
			}

			f, e := os.Open(tt.fpath)
			if e != nil {
				CleanUpTestFile(tt.fpath)
				t.Error(e)
			}

			fi, e := f.Stat()
			if e != nil {
				CleanUpTestFile(tt.fpath)
				t.Error(e)
			}

			if fi.Size() != int64(tt.len) {
				CleanUpTestFile(tt.fpath)
				t.Errorf("New file size is %v, expected %v", fi.Size(), tt.len)
			}

			CleanUpTestFile(tt.fpath)
		})
	}
}
