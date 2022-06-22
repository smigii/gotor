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
	infohash  string
	announce  string
	name      string
	pieceLen  uint64
	pieces    string
	numPieces uint64
	length    uint64
	filelist  *FileList
	bitfield  *utils.Bitfield
}

// ============================================================================
// [G|S]ETTERS ================================================================

func (tor *Torrent) Infohash() string {
	return tor.infohash
}

func (tor *Torrent) Announce() string {
	return tor.announce
}

func (tor *Torrent) Name() string {
	return tor.name
}

func (tor *Torrent) PieceLen() uint64 {
	return tor.pieceLen
}

func (tor *Torrent) NumPieces() uint64 {
	return tor.numPieces
}

func (tor *Torrent) Length() uint64 {
	if tor.filelist != nil {
		return tor.filelist.Length()
	}
	return tor.length
}

func (tor *Torrent) FileList() *FileList {
	return tor.filelist
}

func (tor *Torrent) Piece(idx uint64) (string, error) {
	if idx >= tor.numPieces {
		return "", &TorError{
			msg: fmt.Sprintf("requested piece index [%v], only have [%v]", idx, tor.numPieces),
		}
	}

	offset := idx * 20
	return tor.pieces[offset:20], nil
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewTorrent(fpath string) (*Torrent, error) {

	tor := Torrent{}
	var err error

	fdata, err := os.ReadFile(fpath)
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

	tor.name, err = info.GetString("name")
	if err != nil {
		return nil, err
	}

	tor.pieceLen, err = info.GetUint("piece length")
	if err != nil {
		return nil, err
	}

	// TODO: Read info dictionary manually for SHA1
	// This is rather wasteful
	hasher := sha1.New()
	enc, _ := bencode.Encode(info)
	hasher.Write(enc)
	tor.infohash = string(hasher.Sum(nil))

	// Pieces
	tor.pieces, err = info.GetString("pieces")
	if err != nil {
		return nil, err
	}
	if len(tor.pieces)%20 != 0 {
		return nil, &TorError{
			msg: fmt.Sprintf("'pieces' length must be multiple of 20, got length [%v]", len(tor.pieces)),
		}
	}
	tor.numPieces = uint64(len(tor.pieces) / 20)

	// Length string XOR Files dictionary
	tor.length, err = info.GetUint("length")
	if err != nil {

		// Try 'files'
		files, err := info.GetList("files")
		if err != nil {
			return nil, &TorError{
				msg: fmt.Sprintf("missing keys 'length' and 'files', must have exactly 1"),
			}
		}

		// Read through list of file dictionaries
		tfl, err := extractFileEntries(files, tor.Name())
		if err != nil {
			return nil, err
		}

		tor.filelist = newFileList(tfl, tor.PieceLen())
		// TODO: Create bitfield
	}

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

// extractFileEntries extracts the {path, length} dictionaries from a bencoded
// list.
func extractFileEntries(benlist bencode.List, dirname string) ([]torFileEntry, error) {
	sfl := make([]torFileEntry, 0, 4)

	for _, fEntry := range benlist {
		fDict, ok := fEntry.(bencode.Dict)
		if !ok {
			return nil, &TorError{
				msg: fmt.Sprintf("failed to convert file entry to dictionary\n%v", fEntry),
			}
		}

		fLen, err := fDict.GetUint("length")
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
				return nil, &TorError{
					msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
				}
			}
			strb.WriteString(pathPiece)
			strb.WriteByte('/')
		}
		l := len(strb.String())

		sfl = append(sfl, torFileEntry{
			fpath:  strb.String()[:l-1], // exclude last '/'
			length: fLen,
		})
	}

	return sfl, nil
}

// mkBitfield will look through all the files specified in the torrent and check
// the pieces and their hashes. If a file doesn't exist, the file will be
// created and set to the correct size. If a file exists, but is the wrong
// size, empty bytes will be appended to the correct size. Returns a bitfield
// that represents correct/incorrect piece hashes.
func (tor *Torrent) mkBitfield() {

}

func (tor *Torrent) String() string {
	strb := strings.Builder{}
	prettyHash := hex.EncodeToString([]byte(tor.infohash))

	strb.WriteString("Torrent Info:\n")
	strb.WriteString(fmt.Sprintf("     Name: [%s]\n", tor.name))
	strb.WriteString(fmt.Sprintf(" Announce: [%s]\n", tor.announce))
	strb.WriteString(fmt.Sprintf(" Infohash: [%s]\n", prettyHash))
	plen, units := utils.Bytes4Humans(tor.pieceLen)
	strb.WriteString(fmt.Sprintf("   Pieces: [%v x %v%s]\n", tor.numPieces, plen, units))
	bsize, units := utils.Bytes4Humans(tor.Length())
	strb.WriteString(fmt.Sprintf("   Length: [%.02f %s]\n", bsize, units))

	if tor.filelist != nil {
		strb.WriteString("\nFiles:\n")
		for _, p := range tor.filelist.Files() {
			strb.WriteString(p.fpath)
			strb.WriteByte('\n')
		}
	}

	return strb.String()
}
