package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"gotor/torrent/fileio"

	"gotor/bencode"
	"gotor/utils"
)

// ============================================================================
// ERRORS =====================================================================

type TorError struct{ msg string }

func (te *TorError) Error() string {
	return "tracker error: " + te.msg
}

// ============================================================================
// STRUCTS ====================================================================

type Torrent struct {
	infohash string
	announce string
	fhandle  fileio.FileHandler
}

// ============================================================================
// GETTERS ====================================================================

func (tor *Torrent) Infohash() string {
	return tor.infohash
}

func (tor *Torrent) Announce() string {
	return tor.announce
}

func (tor *Torrent) FileHandler() fileio.FileHandler {
	return tor.fhandle
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewTorrent(paths []string, announce string, pieceLen int64) (*Torrent, error) {

	fentries := make([]fileio.FileEntry, 0, len(paths))

	const mib = 1048576
	const maxBuf = 10 * mib

	npieces := maxBuf / pieceLen
	bufSize := npieces * pieceLen

	buf := make([]byte, bufSize, bufSize)
	buflen := int64(0)
	pieceHashes := strings.Builder{}
	nPieces := int64(0)

	for _, fpath := range paths {

		fptr, e := os.Open(fpath)
		if e != nil {
			return nil, e
		}

		stat, e := fptr.Stat()
		if e != nil {
			return nil, e
		}

		fentries = append(fentries, fileio.MakeFileEntry(fpath, stat.Size()))

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
					nPieces++
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
		nPieces++
	}

	var fh fileio.FileHandler
	if len(paths) == 1 {
		meta, e := fileio.NewTorInfo(paths[0], pieceLen, pieceHashes.String(), fentries)
		if e != nil {
			return nil, e
		}

		fh = fileio.NewSingleFileHandler(meta)
	} else {
		// TODO: implement names and  stuff
		meta, e := fileio.NewTorInfo("TBI", pieceLen, pieceHashes.String(), fentries)
		if e != nil {
			return nil, e
		}

		fh = fileio.NewMultiFileHandler(meta)
	}

	return &Torrent{
		infohash: "",
		announce: announce,
		fhandle:  fh,
	}, nil
}

// FromTorrentFile reads the torrent file specified by torpath and creates a
// new Torrent object.
func FromTorrentFile(torpath string) (*Torrent, error) {

	tor := Torrent{}
	var err error

	fdata, err := os.ReadFile(torpath)
	if err != nil {
		return nil, err
	}

	d, err := bencode.Decode(fdata)
	if err != nil {
		return nil, err
	}

	dict, ok := d.(bencode.Dict)
	if !ok {
		return nil, &TorError{
			msg: "decoded bencoding is not a dictionary",
		}
	}

	tor.announce, err = dict.GetString("announce")
	if err != nil {
		return nil, err
	}

	info, err := dict.GetDict("info")
	if err != nil {
		return nil, err
	}

	// TODO: Read info dictionary manually for SHA1
	// This is rather wasteful
	hasher := sha1.New()
	enc, _ := bencode.Encode(info)
	hasher.Write(enc)
	tor.infohash = string(hasher.Sum(nil))

	fmeta, err := fileio.FromDict(info)
	if err != nil {
		return nil, err
	}

	if fmeta.IsSingle() {
		tor.fhandle = fileio.NewSingleFileHandler(fmeta)
	} else {
		tor.fhandle = fileio.NewMultiFileHandler(fmeta)
	}

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

func (tor *Torrent) String() string {
	meta := tor.fhandle.TorInfo()
	strb := strings.Builder{}
	prettyHash := hex.EncodeToString([]byte(tor.infohash))

	strb.WriteString("Torrent Info:\n")
	strb.WriteString(fmt.Sprintf("     Name: [%s]\n", meta.Name()))
	strb.WriteString(fmt.Sprintf(" Announce: [%s]\n", tor.announce))
	strb.WriteString(fmt.Sprintf(" Infohash: [%s]\n", prettyHash))
	plen, units := utils.Bytes4Humans(meta.PieceLen())
	strb.WriteString(fmt.Sprintf("   Pieces: [%v x %v%s]\n", meta.NumPieces(), plen, units))
	bsize, units := utils.Bytes4Humans(meta.Length())
	strb.WriteString(fmt.Sprintf("   Length: [%.02f %s]\n", bsize, units))

	if !meta.IsSingle() {
		strb.WriteString("\nFiles:\n")
		for _, fe := range meta.Files() {
			strb.WriteString(fe.TorPath())
			strb.WriteByte('\n')
		}
	}

	return strb.String()
}
