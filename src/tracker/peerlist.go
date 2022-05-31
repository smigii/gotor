package tracker

import (
	"encoding/binary"
	"fmt"
	"gotor/bencode"
	"gotor/peer"
	"net"
)

type PeerList []*peer.Peer

// AddPeersFromString
// Reads through a string containing 6 byte peer data defined in BEP_0023
// https://www.bittorrent.org/beps/bep_0023.html
func (p *PeerList) AddPeersFromString(srcString string) error {
	if len(srcString)%6 != 0 {
		return &Error{msg: fmt.Sprintf("compact peer list must be divisible by 6, length = [%v]", len(srcString))}
	}

	nPeers := uint(len(srcString) / 6)

	for i := uint(0); i < nPeers; i++ {
		start := i * 6
		ip := net.IPv4(srcString[start], srcString[start+1], srcString[start+2], srcString[start+3])
		port := binary.BigEndian.Uint16([]byte(srcString[start+4 : start+6]))
		*p = append(*p, peer.NewPeer("", ip, port))
	}

	return nil
}

// AddPeersFromList
// Reads through list of dictionaries defined in BEP_0003
// https://www.bittorrent.org/beps/bep_0003.html#trackers
func (p *PeerList) AddPeersFromList(peerList bencode.List) error {
	for _, v := range peerList {
		peerDict, ok := v.(bencode.Dict)
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
		*p = append(*p, peer.NewPeer(id, net.ParseIP(ip), uint16(port)))
	}
	return nil
}
