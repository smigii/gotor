package utils

import (
	"os"
	"reflect"
	"testing"
)

func TestSegmentData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		segSize uint64
		want    [][]byte
	}{
		{"Single", []byte{0, 1, 2, 3}, 4, [][]byte{{0, 1, 2, 3}}},
		{"Exact", []byte{0, 1, 2, 3, 4, 5}, 2, [][]byte{{0, 1}, {2, 3}, {4, 5}}},
		{"Smaller Last", []byte{0, 1, 2, 3, 4, 5, 6, 7}, 3, [][]byte{{0, 1, 2}, {3, 4, 5}, {6, 7}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SegmentData(tt.data, tt.segSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SegmentData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteEmptyFile(t *testing.T) {

	tests := []struct {
		name  string
		fpath string
		len   uint64
	}{
		{"Single Write", "TestEmptyWrite", writeSize - 100},
		{"Single Write Exact", "TestEmptyWrite", writeSize},
		{"Single File Multi Write", "TestEmptyWrite", writeSize + 100},
		{"Nested File", "TestEmptyWrite/subdir/test", writeSize + 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			e := WriteEmptyFile(tt.fpath, tt.len)
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
