package torrent

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

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
	info     *info.TorInfo
}

// ============================================================================
// GETTERS ====================================================================

func (tor *Torrent) Infohash() string {
	return tor.infohash
}

func (tor *Torrent) Announce() string {
	return tor.announce
}

func (tor *Torrent) Info() *info.TorInfo {
	return tor.info
}

// ============================================================================
// CONSTRUCTOR ================================================================

func NewTorrent(info *info.TorInfo, announce string) (*Torrent, error) {

	// Compute infohash
	bencoded := info.Bencode()
	encoded, err := bencode.Encode(bencoded)
	if err != nil {
		return nil, err
	}
	infohash := utils.SHA1(encoded)

	return &Torrent{
		infohash: infohash,
		announce: announce,
		info:     info,
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

	tor.info = torInfo

	return &tor, nil
}

// ============================================================================
// MISC =======================================================================

func (tor *Torrent) String() string {
	strb := strings.Builder{}
	prettyHash := hex.EncodeToString([]byte(tor.infohash))

	strb.WriteString("Torrent Info:\n")
	strb.WriteString(fmt.Sprintf("     Name: [%s]\n", tor.info.Name()))
	strb.WriteString(fmt.Sprintf(" Announce: [%s]\n", tor.announce))
	strb.WriteString(fmt.Sprintf(" Infohash: [%s]\n", prettyHash))
	plen, units := utils.Bytes4Humans(tor.info.PieceLen())
	strb.WriteString(fmt.Sprintf("   Pieces: [%v x %v %s]\n", tor.info.NumPieces(), plen, units))
	bsize, units := utils.Bytes4Humans(tor.info.Length())
	strb.WriteString(fmt.Sprintf("   Length: [%.02f %s]\n", bsize, units))
	strb.WriteString(fmt.Sprintf("   Length: [%v b]", tor.info.Length()))

	if !tor.info.IsSingle() {
		strb.WriteString("\nFiles:\n")
		for _, fe := range tor.info.Files() {
			size, units2 := utils.Bytes4Humans(fe.Length())
			sizestring := fmt.Sprintf("%v%v", size, units2)
			strb.WriteString(fmt.Sprintf("%8s : %v", sizestring, fe.TorPath()))
			strb.WriteByte('\n')
		}
	}

	return strb.String()
}
