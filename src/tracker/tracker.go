package tracker

import (
	"encoding/binary"
	"fmt"
	"gotor/bencode"
	"gotor/peer"
	"gotor/torrent"
	"gotor/utils"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// ============================================================================
// ERROR ======================================================================

type Error struct{ msg string }

func (e *Error) Error() string {
	return "tracker error: " + e.msg
}

// ============================================================================
// STRUCT =====================================================================

type Resp struct {
	interval    uint64
	minInterval uint64
	warning     string
	tid         string
	seeders     uint64
	leechers    uint64
	peerList    []*peer.Peer
}

// ============================================================================
// GETTERS ====================================================================

func (r Resp) Interval() uint64 {
	return r.interval
}

func (r Resp) MinInterval() uint64 {
	return r.minInterval
}

func (r Resp) Warning() string {
	return r.warning
}

func (r Resp) Tid() string {
	return r.tid
}

func (r Resp) Seeders() uint64 {
	return r.seeders
}

func (r Resp) Leechers() uint64 {
	return r.leechers
}

func (r Resp) Peers() []*peer.Peer {
	return r.peerList
}

// ============================================================================
// FUNK =======================================================================

func Request(tor *torrent.Torrent, port uint16) *http.Request {
	//client := http.Client{}
	req, _ := http.NewRequest("GET", tor.Announce(), nil)
	query := req.URL.Query()
	query.Add("info_hash", tor.Infohash())
	query.Add("peer_id", url.QueryEscape(utils.GotorPeerString+"has93hlsaf83"))
	query.Add("port", fmt.Sprintf("%v", port))
	query.Add("uploaded", fmt.Sprintf("%v", tor.Uploaded()))
	query.Add("downloaded", fmt.Sprintf("%v", tor.Dnloaded()))
	query.Add("left", fmt.Sprintf("%v", tor.Length()))
	query.Add("compact", "1") // For compact peer list (BEP_0023)
	req.URL.RawQuery = query.Encode()
	return req
}

func NewResponse(dict bencode.Dict) (*Resp, error) {
	resp := Resp{}

	// Required fields
	fail, err := dict.GetString("failure reason")
	if err == nil {
		return nil, &Error{msg: fail}
	}

	inter, err := dict.GetUint("interval")
	if err != nil {
		return nil, err
	}
	resp.interval = inter

	seeds, err := dict.GetUint("complete")
	if err != nil {
		return nil, err
	}
	resp.seeders = seeds

	leeches, err := dict.GetUint("incomplete")
	if err != nil {
		return nil, err
	}
	resp.leechers = leeches

	// Peer List
	// 50 peers are returned by default
	resp.peerList = make([]*peer.Peer, 0, 50)

	// Most trackers should be returning compact peer list, try that first
	compactPeerString, err := dict.GetString("peers")
	if err == nil {
		err = resp.AddPeersFromCompactString(compactPeerString)
		if err != nil {
			return nil, err
		}
	} else {
		// Else, tracker probably returned non-compact version
		peerList, err := dict.GetList("peers")
		if err != nil {
			return nil, err
		}
		err = resp.AddPeersFromList(peerList)
		if err != nil {
			return nil, err
		}
	}

	// Optional fields
	warn, err := dict.GetString("warning message")
	if err == nil {
		resp.warning = warn
	}

	minInter, err := dict.GetUint("min interval")
	if err == nil {
		resp.minInterval = minInter
	}

	return &resp, nil
}

// AddPeersFromCompactString
// Reads through a string containing 6 byte peer data defined in BEP_0023
// https://www.bittorrent.org/beps/bep_0023.html
func (r *Resp) AddPeersFromCompactString(peerString string) error {
	if len(peerString)%6 != 0 {
		return &Error{msg: fmt.Sprintf("compact peer list must be divisible by 6, length = [%v]", len(peerString))}
	}
	nPeers := uint(len(peerString) / 6)

	for i := uint(0); i < nPeers; i++ {
		start := i * 6
		ip := net.IPv4(peerString[start], peerString[start+1], peerString[start+2], peerString[start+3])
		port := binary.BigEndian.Uint16([]byte(peerString[start+4 : start+6]))
		r.peerList = append(r.peerList, peer.NewPeer("", ip, port))
	}

	return nil
}

// AddPeersFromList
// Reads through list of dictionaries defined in BEP_0003
// https://www.bittorrent.org/beps/bep_0003.html#trackers
func (r *Resp) AddPeersFromList(peerList bencode.List) error {
	for _, p := range peerList {
		peerDict, ok := p.(bencode.Dict)
		if !ok {
			return &Error{
				msg: fmt.Sprintf("invalid peer list\n[%v]", peerList),
			}
		}
		ip, err := peerDict.GetString("ip")
		if err != nil {
			return err
		}
		port, err := peerDict.GetUint("port")
		if err != nil {
			return err
		}
		id, err := peerDict.GetString("peer id")
		if err != nil {
			return err
		}
		r.peerList = append(r.peerList, peer.NewPeer(id, net.ParseIP(ip), uint16(port)))
	}
	return nil
}

func (r *Resp) Pretty() string {
	strb := strings.Builder{}
	strb.WriteString("Tracker Response:\n")
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Seeders", r.seeders))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Leechers", r.leechers))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Interval", r.interval))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Warning", r.warning))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Tracker ID", r.tid))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Min Interval", r.minInterval))
	strb.WriteString(fmt.Sprintf("Peer List (%v):", len(r.peerList)))
	for i, v := range r.peerList {
		if i%5 == 0 {
			strb.WriteString("\n\t")
		}
		strb.WriteString(fmt.Sprintf("(%v:%v) ", v.Ip(), v.Port()))
	}
	strb.WriteString("\n")
	return strb.String()
}
