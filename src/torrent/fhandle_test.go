package torrent

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestTest(t *testing.T) {

	tor, e := NewTorrent("../../test/multifile.torrent")
	if e != nil {
		t.Error(e)
	}

	fh := NewFileHandler("~/Downloads/", tor)

	fh.Validate()

	x := tor.filelist.GetFiles(0)
	fmt.Println(len(tor.filelist.Files()))
	fmt.Println(len(x))
}

func TestValidate(t *testing.T) {
	//type fields struct {
	//	wd  string
	//	tor *Torrent
	//}
	//tests := []struct {
	//	name   string
	//	fields fields
	//	want   bool
	//}{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		fh := &FileHandler{
	//			wd:  tt.fields.wd,
	//			tor: tt.fields.tor,
	//		}
	//		if got := fh.Validate(); got != tt.want {
	//			t.Errorf("Validate() = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}

func cleanUpTestFile(fpath string) {
	parts := strings.Split(fpath, "/")
	if len(parts) == 0 {
		fmt.Printf("Could not remove test file [%v]\n", fpath)
		return
	}

	e := os.RemoveAll(parts[0]) // Clean up
	if e == nil {
		fmt.Printf("RemoveAll(%v) successful [input %v]\n", parts[0], fpath)
	} else {
		fmt.Printf("RemoveAll(%v) failed [input %v]\n", parts[0], fpath)
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

			e := writeEmptyFile(tt.fpath, tt.len)
			if e != nil {
				cleanUpTestFile(tt.fpath)
				t.Error(e)
			}

			f, e := os.Open(tt.fpath)
			if e != nil {
				cleanUpTestFile(tt.fpath)
				t.Error(e)
			}

			fi, e := f.Stat()
			if e != nil {
				cleanUpTestFile(tt.fpath)
				t.Error(e)
			}

			if fi.Size() != int64(tt.len) {
				cleanUpTestFile(tt.fpath)
				t.Errorf("New file size is %v, expected %v", fi.Size(), tt.len)
			}

			cleanUpTestFile(tt.fpath)
		})
	}
}
