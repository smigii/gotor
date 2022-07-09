package tracker

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"gotor/bencode"
	"gotor/peer"
	"gotor/torrent"
)

// ============================================================================
// ERROR ======================================================================

type Error struct{ msg string }

func (e *Error) Error() string {
	return "tracker error: " + e.msg
}

// ============================================================================
// STRUCT =====================================================================

// State returned by a tracker after a GET request
type State struct {
	interval    uint64
	minInterval uint64
	warning     string
	tid         string
	seeders     uint64
	leechers    uint64
}

// Response is a wrapper for a tracker server response. I didn't want to access
// the peer list through the State object later on, so instead it is returned
// sepereately.
type Response struct {
	State *State
	Peers peer.List
}

// ============================================================================
// GETTERS ====================================================================

func (s State) Interval() uint64 {
	return s.interval
}

func (s State) MinInterval() uint64 {
	return s.minInterval
}

func (s State) Warning() string {
	return s.warning
}

func (s State) Tid() string {
	return s.tid
}

func (s State) Seeders() uint64 {
	return s.seeders
}

func (s State) Leechers() uint64 {
	return s.leechers
}

// ============================================================================
// FUNK =======================================================================

func Get(tor *torrent.Torrent, stats *Stats, port uint16, peerId string) (*Response, error) {
	req := newRequest(tor, stats, port, peerId)
	resp, e := do(req)
	if e != nil {
		return nil, e
	}
	return resp, nil
}

func newRequest(tor *torrent.Torrent, stats *Stats, port uint16, peerId string) *http.Request {
	req, _ := http.NewRequest("GET", tor.Announce(), nil)
	query := req.URL.Query()
	query.Add("info_hash", tor.Infohash())
	query.Add("peer_id", url.QueryEscape(peerId))
	query.Add("port", fmt.Sprintf("%v", port))
	query.Add("uploaded", fmt.Sprintf("%v", stats.Uploaded()))
	query.Add("downloaded", fmt.Sprintf("%v", stats.Dnloaded()))
	query.Add("left", fmt.Sprintf("%v", stats.Left()))
	query.Add("compact", "1") // For compact peer list (BEP_0023)
	req.URL.RawQuery = query.Encode()
	return req
}

func do(req *http.Request) (*Response, error) {
	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ben, err := bencode.Decode(body)
	if err != nil {
		return nil, err
	}

	dict, ok := ben.(bencode.Dict)
	if !ok {
		return nil, fmt.Errorf("response not a bencoded dictionary\n%v", body)
	}

	tresp, err := newResponse(dict)
	if err != nil {
		return nil, err
	} else {
		return tresp, nil
	}
}

func newResponse(dict bencode.Dict) (*Response, error) {
	state := State{}
	resp := Response{
		State: &state,
		Peers: nil,
	}

	// Required fields --------------------------------------------------------
	fail, err := dict.GetString("failure reason")
	if err == nil {
		return nil, &Error{msg: fail}
	}

	inter, err := dict.GetUint("interval")
	if err != nil {
		return nil, err
	}
	state.interval = inter

	seeds, err := dict.GetUint("complete")
	if err != nil {
		return nil, err
	}
	state.seeders = seeds

	leeches, err := dict.GetUint("incomplete")
	if err != nil {
		return nil, err
	}
	state.leechers = leeches

	// Get the peer list ------------------------------------------------------
	// Interface which will be used to extract peers (either a string or list)
	var src peer.ListSource

	// Most trackers should be returning compact peer list, try that first
	compactPeerString, err := dict.GetString("peers")
	if err == nil {
		src = stringSource(compactPeerString)
	} else {
		// Else, tracker probably returned non-compact version
		list, err := dict.GetList("peers")
		if err != nil {
			// Tracker returned some hot trash
			return nil, err
		}
		src = listSource(list)
	}
	resp.Peers, err = src.GetPeers()
	if err != nil {
		return nil, err
	}

	// Optional fields --------------------------------------------------------
	warn, err := dict.GetString("warning message")
	if err == nil {
		state.warning = warn
	}

	minInter, err := dict.GetUint("min interval")
	if err == nil {
		state.minInterval = minInter
	}

	return &resp, nil
}

// ============================================================================
// IMPLEMENT INTERFACE ========================================================

// Define 2 new ways to add peers to a peer list using the interface from
// list.go; using strings and using bencoded lists

type stringSource string

func (s stringSource) GetPeers() (peer.List, error) {
	if len(s)%6 != 0 {
		return nil, &Error{msg: fmt.Sprintf("compact peer list must be divisible by 6, length = [%v]", len(s))}
	}

	nPeers := uint(len(s) / 6)
	peerList := make(peer.List, 0, nPeers)

	for i := uint(0); i < nPeers; i++ {
		start := i * 6
		ip := net.IPv4(s[start], s[start+1], s[start+2], s[start+3])
		port := binary.BigEndian.Uint16([]byte(s[start+4 : start+6]))
		peerList = append(peerList, peer.MakePeer("", ip, port))
	}

	return peerList, nil
}

type listSource bencode.List

func (l listSource) GetPeers() (peer.List, error) {
	peerList := make(peer.List, 0, len(l))
	for _, v := range l {
		peerDict, ok := v.(bencode.Dict)
		if !ok {
			return nil, &Error{
				msg: fmt.Sprintf("invalid peer list\n[%v]", peerList),
			}
		}
		ip, err := peerDict.GetString("ip")
		if err != nil {
			return nil, err
		}
		port, err := peerDict.GetUint("port")
		if err != nil {
			return nil, err
		}
		id, err := peerDict.GetString("peer id")
		if err != nil {
			return nil, err
		}
		peerList = append(peerList, peer.MakePeer(id, net.ParseIP(ip), uint16(port)))
	}
	return peerList, nil
}

// ============================================================================
// STRING =====================================================================

func (s *State) String() string {
	strb := strings.Builder{}
	strb.WriteString("Tracker State:\n")
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Seeders", s.seeders))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Leechers", s.leechers))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Interval", s.interval))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Warning", s.warning))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Tracker ID", s.tid))
	strb.WriteString(fmt.Sprintf("%12s: [%v]\n", "Min Interval", s.minInterval))
	return strb.String()
}

// Temporary
func (r *Response) String() string {
	strb := strings.Builder{}
	strb.WriteString(r.State.String())
	strb.WriteString(fmt.Sprintf("Peer List (%v):", len(r.Peers)))
	for i, v := range r.Peers {
		if i%5 == 0 {
			strb.WriteString("\n\t")
		}
		strb.WriteString(fmt.Sprintf("(%v:%v) ", v.Ip(), v.Port()))
	}
	return strb.String()
}
