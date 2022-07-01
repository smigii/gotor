package fileio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gotor/bencode"
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
	length    int64       `info:"length"`
	files     []FileEntry `info:"files"`
	isSingle  bool        // Is this a single-file or multi-file torrent?
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

func (ti *TorInfo) Files() []FileEntry {
	return ti.files
}

func (ti *TorInfo) IsSingle() bool {
	return ti.isSingle
}

// ============================================================================
// CONSTRUCTOR ================================================================

func CreateTorInfo(paths []string, workingDir string, name string, pieceLen int64) (*TorInfo, error) {
	fentries := make([]FileEntry, 0, len(paths))

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

		fentry := MakeFileEntry(fpath, stat.Size())
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
	}

	// Process anything remaining in the buffer
	pieces := utils.SegmentData(buf[:buflen], pieceLen)
	for _, piece := range pieces {
		pieceHash := utils.SHA1(piece)
		pieceHashes.WriteString(pieceHash)
	}

	return NewTorInfo(name, pieceLen, pieceHashes.String(), fentries)
}

func NewTorInfo(name string, pieceLen int64, hashes string, files []FileEntry) (*TorInfo, error) {

	if len(hashes)%20 != 0 {
		return nil, errors.New("hashes must be multiple of 20")
	}

	length := int64(0)

	for _, fentry := range files {
		length += fentry.Length()
	}

	return &TorInfo{
		name:      name,
		pieceLen:  pieceLen,
		hashes:    hashes,
		numPieces: int64(len(hashes) / 20),
		length:    length,
		files:     files,
		isSingle:  len(files) == 1,
	}, nil

}

// FromDict creates and returns a new *TorInfo from the info dictionary of a
// torrent file.
func FromDict(info bencode.Dict, workingDir string) (*TorInfo, error) {
	fdata := TorInfo{}
	var err error

	fdata.name, err = info.GetString("name")
	if err != nil {
		return nil, err
	}

	fdata.pieceLen, err = info.GetInt("piece length")
	if err != nil {
		return nil, err
	}

	fdata.hashes, err = info.GetString("pieces")
	if err != nil {
		return nil, err
	}
	if len(fdata.hashes)%20 != 0 {
		return nil, &FileMetaError{
			msg: fmt.Sprintf("'pieces' length must be multiple of 20, got length [%v]", len(fdata.hashes)),
		}
	}
	fdata.numPieces = int64(len(fdata.hashes) / 20)

	fdata.length, err = info.GetInt("length")
	if err == nil {
		fdata.isSingle = true

		fentry := MakeFileEntry(fdata.Name(), fdata.Length())
		localPath := path.Join(workingDir, fentry.TorPath())
		fentry.SetLocalPath(localPath)
		fdata.files = []FileEntry{fentry}

	} else {
		fdata.isSingle = false

		// Try 'files'
		files, err := info.GetList("files")
		if err != nil {
			return nil, &FileMetaError{
				msg: fmt.Sprintf("missing keys 'length' and 'files', must have exactly 1"),
			}
		}

		// Read through list of file dictionaries
		fdata.files, err = extractFileEntries(files, fdata.name)
		if err != nil {
			return nil, err
		}
	}

	return &fdata, nil
}

// ============================================================================
// FUNC =======================================================================

func (ti *TorInfo) PieceHash(idx int64) (string, error) {
	if idx >= ti.numPieces {
		return "", &FileMetaError{
			msg: fmt.Sprintf("requested piece index [%v], max is [%v]", idx, ti.numPieces-1),
		}
	}

	offset := idx * 20
	return ti.hashes[offset : offset+20], nil
}

// extractFileEntries extracts the {path, length} dictionaries from a bencoded
// list.
func extractFileEntries(benlist bencode.List, dirname string) ([]FileEntry, error) {
	sfl := make([]FileEntry, 0, 4)

	for _, fEntry := range benlist {
		fDict, ok := fEntry.(bencode.Dict)
		if !ok {
			return nil, &FileMetaError{
				msg: fmt.Sprintf("failed to convert file entry to dictionary\n%v", fEntry),
			}
		}

		fLen, err := fDict.GetInt("length")
		if err != nil {
			return nil, err
		}

		fPathList, err := fDict.GetList("path")
		if err != nil {
			return nil, err
		}

		// Read through list of path strings
		strb := strings.Builder{}

		// Write the directory name
		strb.WriteString(dirname)
		strb.WriteByte('/')

		for _, fPathEntry := range fPathList {
			pathPiece, ok := fPathEntry.(string)
			if !ok {
				return nil, &FileMetaError{
					msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
				}
			}
			strb.WriteString(pathPiece)
			strb.WriteByte('/')
		}
		l := len(strb.String())

		// exclude last '/'
		sfl = append(sfl, MakeFileEntry(strb.String()[:l-1], fLen))
	}

	return sfl, nil
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
