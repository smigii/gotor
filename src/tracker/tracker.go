package tracker

import (
	"fmt"
	"gotor/bencode"
	"gotor/torrent"
	"io/ioutil"
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

type State struct {
	interval    uint64
	minInterval uint64
	warning     string
	tid         string
	seeders     uint64
	leechers    uint64
}

type Response struct {
	State *State
	Peers PeerList
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

func NewRequest(tor *torrent.Torrent, port uint16, peerId string) *http.Request {
	req, _ := http.NewRequest("GET", tor.Announce(), nil)
	query := req.URL.Query()
	query.Add("info_hash", tor.Infohash())
	query.Add("peer_id", url.QueryEscape(peerId))
	query.Add("port", fmt.Sprintf("%v", port))
	query.Add("uploaded", fmt.Sprintf("%v", tor.Uploaded()))
	query.Add("downloaded", fmt.Sprintf("%v", tor.Dnloaded()))
	query.Add("left", fmt.Sprintf("%v", tor.Length()))
	query.Add("compact", "1") // For compact peer list (BEP_0023)
	req.URL.RawQuery = query.Encode()
	return req
}

func Do(req *http.Request) (*Response, error) {
	client := http.Client{}
	resp, err := client.Do(req)

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

	tresp, err := NewResponse(dict)
	if err != nil {
		return nil, err
	} else {
		return tresp, nil
	}
}

func NewResponse(dict bencode.Dict) (*Response, error) {
	state := State{}

	// Required fields
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

	// Peer List
	// 50 peers are returned by default
	peerList := make(PeerList, 0, 50)

	// Most trackers should be returning compact peer list, try that first
	compactPeerString, err := dict.GetString("peers")
	if err == nil {
		err = peerList.AddPeersFromString(compactPeerString)
		if err != nil {
			return nil, err
		}
	} else {
		// Else, tracker probably returned non-compact version
		list, err := dict.GetList("peers")
		if err != nil {
			return nil, err
		}
		err = peerList.AddPeersFromList(list)
		if err != nil {
			return nil, err
		}
	}

	// Optional fields
	warn, err := dict.GetString("warning message")
	if err == nil {
		state.warning = warn
	}

	minInter, err := dict.GetUint("min interval")
	if err == nil {
		state.minInterval = minInter
	}

	resp := Response{
		State: &state,
		Peers: peerList,
	}
	return &resp, nil
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
