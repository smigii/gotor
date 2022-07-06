package info

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gotor/bencode"
	"gotor/torrent/filesd"
	"gotor/utils"
)

// ============================================================================
// ERRORS =====================================================================

type FileMetaError struct{ msg string }

func (fme *FileMetaError) Error() string {
	return "tracker error: " + fme.msg
}

// ============================================================================
// STRUCTS ====================================================================

// TorInfo holds the metadata found in a torrent's info dictionary.
type TorInfo struct {
	name      string `info:"name"`
	pieceLen  int64  `info:"piece length"`
	hashes    string `info:"pieces"`
	numPieces int64
	length    int64           `info:"length"`
	flist     filesd.FileList `info:"files"`
	isSingle  bool            // Is this a single-file or multi-file torrent?

	pm PieceMap
}

// ============================================================================
// GETTERS ====================================================================

func (ti *TorInfo) Name() string {
	return ti.name
}

func (ti *TorInfo) PieceLen() int64 {
	return ti.pieceLen
}

func (ti *TorInfo) Hashes() string {
	return ti.hashes
}

func (ti *TorInfo) NumPieces() int64 {
	return ti.numPieces
}

func (ti *TorInfo) Length() int64 {
	return ti.length
}

func (ti *TorInfo) Files() filesd.FileList {
	return ti.flist
}

func (ti *TorInfo) IsSingle() bool {
	return ti.isSingle
}

// ============================================================================
// CONSTRUCTOR ================================================================

func CreateTorInfo(paths []string, workingDir string, name string, pieceLen int64) (*TorInfo, error) {
	fentries := make([]filesd.EntryBase, 0, len(paths))

	const mib = 1048576
	const maxBuf = 5 * mib

	npieces := maxBuf / pieceLen
	bufSize := npieces * pieceLen

	buf := make([]byte, bufSize, bufSize)
	buflen := int64(0)
	pieceHashes := strings.Builder{}

	// Multifile torrents don't include the base directory in the files dictionary.
	if len(paths) > 1 {
		workingDir = filepath.Join(workingDir, name)
	}

	for _, fpath := range paths {

		localPath := filepath.Join(workingDir, fpath)

		fptr, e := os.Open(localPath)
		if e != nil {
			return nil, e
		}

		stat, e := fptr.Stat()
		if e != nil {
			return nil, e
		}

		fentry := filesd.MakeFileEntry(fpath, stat.Size())
		fentry.SetLocalPath(localPath)
		fentries = append(fentries, fentry)

		for {
			// Fill buffer with data
			n, e := fptr.Read(buf[buflen:])
			if e != nil {
				if e != io.EOF {
					return nil, e
				}
				// If we reached end of file, break and start processing
				// next file.
				break
			}
			buflen += int64(n)

			// If the buffer is now full, process the pieces.
			// then "clear" the buffer and loop again.
			if buflen == bufSize {
				pieces := utils.SegmentData(buf, pieceLen)
				for _, piece := range pieces {
					pieceHash := utils.SHA1(piece)
					pieceHashes.WriteString(pieceHash)
				}
				buflen = 0
			}
		}

		err := fptr.Close()
		if err != nil {
			return nil, err
		}
	}

	// Process anything remaining in the buffer
	pieces := utils.SegmentData(buf[:buflen], pieceLen)
	for _, piece := range pieces {
		pieceHash := utils.SHA1(piece)
		pieceHashes.WriteString(pieceHash)
	}

	return NewTorInfo(name, pieceLen, pieceHashes.String(), fentries)
}

func NewTorInfo(name string, pieceLen int64, hashes string, files []filesd.EntryBase) (*TorInfo, error) {

	if len(hashes)%20 != 0 {
		return nil, errors.New("hashes must be multiple of 20")
	}

	length := int64(0)

	for _, fentry := range files {
		length += fentry.Length()
	}

	flist := filesd.MakeFileList(files, pieceLen)

	npieces := int64(len(hashes) / 20)
	pm, e := MakePieceMap(flist, npieces, pieceLen, length)
	if e != nil {
		return nil, e
	}

	return &TorInfo{
		name:      name,
		pieceLen:  pieceLen,
		hashes:    hashes,
		numPieces: int64(len(hashes) / 20),
		length:    length,
		flist:     flist,
		isSingle:  len(files) == 1,
		pm:        pm,
	}, nil

}

// FromDict creates and returns a new *TorInfo from the info dictionary of a
// torrent file.
func FromDict(info bencode.Dict, workingDir string) (*TorInfo, error) {

	name, err := info.GetString("name")
	if err != nil {
		return nil, err
	}

	pieceLen, err := info.GetInt("piece length")
	if err != nil {
		return nil, err
	}

	hashes, err := info.GetString("pieces")
	if err != nil {
		return nil, err
	}

	var fentries []filesd.EntryBase
	length, err := info.GetInt("length")
	if err == nil {

		fentry := filesd.MakeFileEntry(name, length)
		localPath := filepath.Join(workingDir, fentry.TorPath())
		fentry.SetLocalPath(localPath)
		fentries = []filesd.EntryBase{fentry}

	} else {

		// Try 'files'
		files, err := info.GetList("files")
		if err != nil {
			return nil, &FileMetaError{
				msg: fmt.Sprintf("missing keys 'length' and 'files', must have exactly 1"),
			}
		}

		// Read through list of file dictionaries
		fentries, err = filesd.FromBenList(files)
		if err != nil {
			return nil, err
		}

		// Modify local path
		workingDir = filepath.Join(workingDir, name) // Torrent paths don't include base dir
		for i, fe := range fentries {
			localPath := filepath.Join(workingDir, fe.TorPath())
			fentries[i].SetLocalPath(localPath)
		}
	}

	return NewTorInfo(name, pieceLen, hashes, fentries)
}

// ============================================================================
// FUNC =======================================================================

func (ti *TorInfo) PieceHash(idx int64) string {
	offset := idx * 20
	return ti.hashes[offset : offset+20]
}

func (ti *TorInfo) Bencode() bencode.Dict {
	d := make(bencode.Dict)

	d["name"] = ti.Name()
	d["piece length"] = ti.PieceLen()
	d["pieces"] = ti.Hashes()

	// This is present in transmission torrents, so for now we will keep
	// this in here for testing purposes (comparing known vs computed
	// infohashes
	d["private"] = int64(0)

	if ti.IsSingle() {
		d["length"] = ti.Length()
	} else {
		list := make(bencode.List, 0, len(ti.Files()))
		for _, fentry := range ti.Files() {
			fdict := fentry.Bencode()
			list = append(list, fdict)
		}
		d["files"] = list
	}

	return d
}

func (ti *TorInfo) PieceLookup(index int64) ([]PieceLocation, error) {

	if index < 0 {
		return nil, errors.New("negative index")
	}

	if index >= ti.numPieces {
		return nil, errors.New("index out of bounds")
	}

	return ti.pm[index], nil
}
