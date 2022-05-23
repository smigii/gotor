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
	files     []FileEntry

	uploaded uint64
	dnloaded uint64
}

type FileEntry struct {
	length uint64
	path   []string
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
	return tor.length
}

func (tor *Torrent) Files() []FileEntry {
	return tor.files
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

func (tor *Torrent) Uploaded() uint64 {
	return tor.uploaded
}

func (tor *Torrent) SetUploaded(uploaded uint64) {
	tor.uploaded = uploaded
}

func (tor *Torrent) IncUploaded(amnt uint64) {
	tor.uploaded += amnt
}

func (tor *Torrent) Dnloaded() uint64 {
	return tor.dnloaded
}

func (tor *Torrent) SetDnloaded(dnloaded uint64) {
	tor.dnloaded = dnloaded
}

func (tor *Torrent) IncDownloaded(amnt uint64) {
	tor.dnloaded += amnt
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewTorrent(path string) (*Torrent, error) {

	tor := Torrent{}
	var err error

	fdata, err := os.ReadFile(path)
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
		tor.files = make([]FileEntry, 0, 8)
		for _, fEntry := range files {
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
			tor.length += fLen
			fPathList, err := fDict.GetList("path")
			if err != nil {
				return nil, err
			}

			// Read through list of path strings
			pathPieces := make([]string, 0, 2)
			for _, fPathEntry := range fPathList {
				pathPiece, ok := fPathEntry.(string)
				if !ok {
					return nil, &TorError{
						msg: fmt.Sprintf("file entry contains invalid path [%v]", fEntry),
					}
				}
				pathPieces = append(pathPieces, pathPiece)
			}

			tor.files = append(tor.files, FileEntry{
				length: fLen,
				path:   pathPieces,
			})
		}
	}

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

func (tor *Torrent) QuickStats() string {
	builder := strings.Builder{}
	prettyHash := hex.EncodeToString([]byte(tor.infohash))
	builder.WriteString(fmt.Sprintf("     Name: [%s]\n", tor.name))
	builder.WriteString(fmt.Sprintf(" Announce: [%s]\n", tor.announce))
	builder.WriteString(fmt.Sprintf(" Infohash: [%s]\n", prettyHash))
	bsize, units := utils.Bytes4Humans(tor.length)
	builder.WriteString(fmt.Sprintf("   Length: [%.02f %v]\n", bsize, units))

	if tor.files != nil {
		builder.WriteString("\nFiles:\n")
		for _, p := range tor.files {
			builder.WriteString(strings.Join(p.path, "/"))
			builder.WriteByte('\n')
		}
	}

	return builder.String()
}
