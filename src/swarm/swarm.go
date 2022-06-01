package swarm

import (
	"gotor/peer"
	"gotor/torrent"
	"gotor/tracker"
	"gotor/utils"
	"log"
	"net"
	"strings"
)

// ============================================================================
// STRUCTS ====================================================================

type Swarm struct {
	State *tracker.State
	Stats *tracker.Stats
	Peers peer.List
	Tor   *torrent.Torrent
	Id    string
	Port  uint16
}

// ============================================================================
// FUNK =======================================================================

func NewSwarm(opts *utils.Opts) (*Swarm, error) {
	var err error

	swarm := Swarm{}
	swarm.Id = utils.NewPeerId()
	swarm.Port = opts.Lport()

	// Read torrent file
	log.Printf("reading torrent file [%v]\n", opts.Input())
	swarm.Tor, err = torrent.NewTorrent(opts.Input())
	if err != nil {
		return nil, err
	}

	// TODO: Check opts.output and see how much is really done
	swarm.Stats = tracker.NewStats(0, 0, swarm.Tor.Length())

	// Make first contact with tracker
	log.Printf("sending get to tracker [%v]\n", swarm.Tor.Announce())
	resp, err := tracker.Get(swarm.Tor, swarm.Stats, swarm.Port, swarm.Id)
	if err != nil {
		return nil, err
	}

	swarm.State = resp.State
	swarm.Peers = resp.Peers

	return &swarm, nil
}

func (s *Swarm) Start() error {

	// Start peer Goroutines
	for _, p := range s.Peers {
		go s.HandlePeer(p)
	}

	return nil
}

func (s *Swarm) HandlePeer(peer *peer.Peer) {
	e := s.Bootstrap(peer)
	if e != nil {
		// do something?
	}

	//for {
	//
	//}

}

// Bootstrap
// Create a TCP connection with the peer, then send
// the BitTorrent handshake.
func (s *Swarm) Bootstrap(peer *peer.Peer) error {
	var err error
	peer.Conn, err = net.Dial("tcp", peer.Addr())

	if err != nil {
		return err
	}

	hs := MakeHandshake(s.Tor.Infohash(), s.Id)
	_, err = peer.Conn.Write(hs)
	if err != nil {
		return nil
	}

	return nil
}

func (s *Swarm) String() string {
	strb := strings.Builder{}
	strb.WriteString(s.Tor.String())
	strb.WriteByte('\n')
	strb.WriteString(s.State.String())
	strb.WriteByte('\n')
	strb.WriteString(s.Peers.String())
	return strb.String()
}
