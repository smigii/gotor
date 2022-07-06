package fileio

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"gotor/torrent/filesd"
	"gotor/torrent/info"
	"gotor/utils"
)

// ============================================================================
// ERRORS =====================================================================

type PathError struct {
	fpath string
}

func (pe *PathError) Error() string {
	return fmt.Sprintf("path [%v] does not exist", pe.fpath)
}

// ============================================================================
// STRUCTS ====================================================================

type lockedFp struct {
	fp   *os.File
	lock sync.RWMutex
}

type FileIO struct {
	lfps map[string]*lockedFp
}

// ============================================================================
// FUNC =======================================================================

func NewFileIO() *FileIO {
	return &FileIO{
		lfps: make(map[string]*lockedFp),
	}
}

func newLockedFp(fp *os.File) *lockedFp {
	return &lockedFp{
		fp:   fp,
		lock: sync.RWMutex{},
	}
}

// OCAT will call utils.OCAT to open/create/append/truncate a file
// to the appropriate size.
func (fio *FileIO) OCAT(fpath string, length int64) error {
	f, e := utils.OCAT(fpath, length)
	if e != nil {
		return e
	} else {
		fio.lfps[fpath] = newLockedFp(f)
		return nil
	}
}

func (fio *FileIO) OCATAll(files filesd.FileList) error {

	for _, fe := range files {
		e := fio.OCAT(fe.LocalPath(), fe.Length())
		if e != nil {
			return e
		}
	}
	return nil

}

func (fio *FileIO) Move(fromPath string, toPath string) error {
	lfp, ok := fio.lfps[fromPath]
	if ok {
		lfp.lock.Lock()
		defer lfp.lock.Unlock()
		e := os.Rename(fromPath, toPath)
		return e
	} else {
		return &PathError{fpath: fromPath}
	}
}

func (fio *FileIO) Close(fpath string) error {
	lfp, ok := fio.lfps[fpath]
	if ok {
		lfp.lock.Lock()
		defer lfp.lock.Unlock()
		e := lfp.fp.Close()
		return e
	} else {
		return &PathError{fpath: fpath}
	}
}

func (fio *FileIO) CloseAll() error {
	var e error

	for _, lfp := range fio.lfps {
		func() {
			lfp.lock.Lock()
			defer lfp.lock.Unlock()
			e = lfp.fp.Close()
		}()

		if e != nil {
			return e
		}
	}

	return nil
}

func (fio *FileIO) write(fpath string, seekAmnt int64, data []byte) (int64, error) {
	lfp, ok := fio.lfps[fpath]
	if ok {
		lfp.lock.Lock()
		defer lfp.lock.Unlock()
		n, e := lfp.fp.WriteAt(data, seekAmnt)
		return int64(n), e
	} else {
		return 0, &PathError{fpath: fpath}
	}
}

func (fio *FileIO) read(fpath string, seekAmnt int64, buf []byte) (int64, error) {
	lfp, ok := fio.lfps[fpath]
	if ok {
		lfp.lock.RLock()
		defer lfp.lock.RUnlock()
		n, e := lfp.fp.ReadAt(buf, seekAmnt)
		return int64(n), e
	} else {
		return 0, &PathError{fpath: fpath}
	}
}

func (fio *FileIO) ReadPiece(index int64, torInfo *info.TorInfo, buf []byte) (int64, error) {

	plocs, e := torInfo.PieceLookup(index)
	if e != nil {
		return 0, e
	}

	offset := int64(0)
	for _, ploc := range plocs {
		subbuf := buf[offset : offset+ploc.Loc.ReadAmnt]
		n, e := fio.read(ploc.Entry.LocalPath(), ploc.Loc.SeekAmnt, subbuf)
		if e != nil {
			return 0, e
		}
		offset += n
	}

	return offset, nil
}

func (fio *FileIO) WritePiece(index int64, torInfo *info.TorInfo, data []byte) (int64, error) {

	plocs, e := torInfo.PieceLookup(index)
	if e != nil {
		return 0, e
	}

	knownHash := torInfo.PieceHash(index)

	if utils.SHA1(data) != knownHash {
		return 0, errors.New("invalid hash, refusing write")
	}

	offset := int64(0)
	for _, ploc := range plocs {
		subbuf := data[offset : offset+ploc.Loc.ReadAmnt]
		n, e := fio.write(ploc.Entry.LocalPath(), ploc.Loc.SeekAmnt, subbuf)
		if e != nil {
			return 0, e
		}
		offset += n
	}

	return offset, nil
}
