package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"gotor/bencode"
	"gotor/utils"
	"os"
	"strings"
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
	fhandle  FileHandler
	bitfield *utils.Bitfield
}

// ============================================================================
// GETTERS ====================================================================

func (tor *Torrent) Infohash() string {
	return tor.infohash
}

func (tor *Torrent) Announce() string {
	return tor.announce
}

func (tor *Torrent) FileHandler() FileHandler {
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

	fmeta, err := newTorFileMeta(info)
	if err != nil {
		return nil, err
	}

	if fmeta.isSingle {

	} else {
		tor.fhandle = newFileList(fmeta)
	}

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

// mkBitfield will look through all the files specified in the torrent and check
// the pieces and their hashes. If a file doesn't exist, the file will be
// created and set to the correct size. If a file exists, but is the wrong
// size, empty bytes will be appended to the correct size. Returns a bitfield
// that represents correct/incorrect piece hashes.
func (tor *Torrent) mkBitfield() {

}

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
		for _, p := range meta.Files() {
			strb.WriteString(p.fpath)
			strb.WriteByte('\n')
		}
	}

	return strb.String()
}
