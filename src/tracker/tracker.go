package tracker

import (
	"fmt"
	"gotor/bencode"
	"gotor/torrent"
	"io/ioutil"
	"net/http"
	"net/url"
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
	peers       []Peer
}

type Peer struct {
	id   string
	ip   string
	port uint16
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

func (r Resp) Peers() []Peer {
	return r.peers
}

// ============================================================================
// FUNK =======================================================================

func Request(tor *torrent.Torrent) (*Resp, error) {
	// https://torrent.ubuntu.com/announce?info_hash=%F0%9C%8D%08%84Y%00%88%F4%00N%01%0A%92%8F%8Bax%C2%FD&peer_id=shf74nfdhas93hlsaf83&port=0&uploaded=0&downloaded=0&left=0
	client := http.Client{}
	req, _ := http.NewRequest("GET", tor.Announce(), nil)
	query := req.URL.Query()
	query.Add("info_hash", tor.Infohash())
	query.Add("peer_id", url.QueryEscape("G0T0R-fdhas93hlsaf83"))
	query.Add("port", "30666")
	query.Add("uploaded", fmt.Sprintf("%v", tor.Uploaded()))
	query.Add("downloaded", fmt.Sprintf("%v", tor.Dnloaded()))
	query.Add("left", fmt.Sprintf("%v", tor.Length()))
	req.URL.RawQuery = query.Encode()
	//fmt.Println("\nFull URL: ", req.URL)
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
		return nil, &Error{
			msg: fmt.Sprintf("response not a bencoded dictionary\n%v", body),
		}
	}

	tresp, err := newResponse(dict)
	if err != nil {
		return nil, err
	} else {
		return tresp, nil
	}
}

func newResponse(dict bencode.Dict) (*Resp, error) {
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
	peerList, err := dict.GetList("peers")
	if err != nil {
		return nil, err
	}
	// 50 peers are returned by default
	resp.peers = make([]Peer, 0, 50)
	for _, p := range peerList {
		peerDict, ok := p.(bencode.Dict)
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

		resp.peers = append(resp.peers, Peer{
			id:   id,
			ip:   ip,
			port: uint16(port),
		})
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
