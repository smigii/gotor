package torrent

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"gotor/torrent/fileio"
	"gotor/torrent/info"

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

func NewTorrent(info *info.TorInfo, announce string) (*Torrent, error) {
	// Make FileHandler
	var fh fileio.FileHandler
	if len(info.Files()) == 1 {
		fh = fileio.NewSingleFileHandler(info)
	} else {
		fh = fileio.NewMultiFileHandler(info)
	}

	// Compute infohash
	bencoded := fh.TorInfo().Bencode()
	encoded, err := bencode.Encode(bencoded)
	if err != nil {
		return nil, err
	}
	infohash := utils.SHA1(encoded)

	return &Torrent{
		infohash: infohash,
		announce: announce,
		fhandle:  fh,
	}, nil
}

// FromTorrentFile reads the torrent file specified by torpath and creates a
// new Torrent object.
func FromTorrentFile(torpath string, workingDir string) (*Torrent, error) {

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
		return nil, &TorError{msg: "decoded bencoding is not a dictionary"}
	}

	tor.announce, err = dict.GetString("announce")
	if err != nil {
		return nil, err
	}

	infodict, err := dict.GetDict("info")
	if err != nil {
		return nil, err
	}

	enc, _ := bencode.Encode(infodict)
	tor.infohash = utils.SHA1(enc)

	torInfo, err := info.FromDict(infodict, workingDir)
	if err != nil {
		return nil, err
	}

	if torInfo.IsSingle() {
		tor.fhandle = fileio.NewSingleFileHandler(torInfo)
	} else {
		tor.fhandle = fileio.NewMultiFileHandler(torInfo)
	}

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

func (tor *Torrent) String() string {
	torInfo := tor.fhandle.TorInfo()
	strb := strings.Builder{}
	prettyHash := hex.EncodeToString([]byte(tor.infohash))

	strb.WriteString("Torrent Info:\n")
	strb.WriteString(fmt.Sprintf("     Name: [%s]\n", torInfo.Name()))
	strb.WriteString(fmt.Sprintf(" Announce: [%s]\n", tor.announce))
	strb.WriteString(fmt.Sprintf(" Infohash: [%s]\n", prettyHash))
	plen, units := utils.Bytes4Humans(torInfo.PieceLen())
	strb.WriteString(fmt.Sprintf("   Pieces: [%v x %v%s]\n", torInfo.NumPieces(), plen, units))
	bsize, units := utils.Bytes4Humans(torInfo.Length())
	strb.WriteString(fmt.Sprintf("   Length: [%.02f %s]\n", bsize, units))

	if !torInfo.IsSingle() {
		strb.WriteString("\nFiles:\n")
		for _, fe := range torInfo.Files() {
			strb.WriteString(fe.TorPath())
			strb.WriteByte('\n')
		}
	}

	return strb.String()
}
