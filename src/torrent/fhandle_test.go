package torrent

import (
	"fmt"
	"testing"
)

func TestTest(t *testing.T) {

	tor, e := NewTorrent("../../test/multifile.torrent")
	if e != nil {
		t.Error(e)
	}

	fh := NewFileHandler("~/Downloads/", tor)

	fh.Validate()

	//e = writeEmptyFile("/home/smigii/Downloads/test/test", 698351616)
	//if e != nil {
	//	t.Error(e)
	//}

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
