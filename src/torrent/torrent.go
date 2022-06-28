package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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

func NewTorrent(fpath string) (*Torrent, error) {

	tor := Torrent{}
	var err error

	f, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	d, err := bencode.Decode(f)
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

	fmeta, err := fileio.NewFileMeta(info)
	if err != nil {
		return nil, err
	}

	if fmeta.IsSingle() {
		tor.fhandle, err = fileio.NewSingleFileHandler(fmeta)
		if err != nil {
			panic(err)
		}
	} else {
		tor.fhandle, err = fileio.NewMultiFileHandler(fmeta)
		if err != nil {
			panic(err)
		}
	}

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

func (tor *Torrent) String() string {
	meta := tor.fhandle.FileMeta()
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
			strb.WriteString(fe.Path())
			strb.WriteByte('\n')
		}
	}

	return strb.String()
}
