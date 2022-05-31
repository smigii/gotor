package swarm

import (
	"fmt"
	"gotor/peer"
	"gotor/torrent"
	"gotor/tracker"
	"gotor/utils"
	"net"
)

type Swarm struct {
	Tor     *torrent.Torrent
	Tracker *tracker.Resp
	//peers   []*peer.Peer
	Id string
}

func NewSwarm(path string, port uint16) (*Swarm, error) {
	swarm := Swarm{}
	swarm.Id = utils.NewPeerId()
	var err error

	// Read torrent file
	swarm.Tor, err = torrent.NewTorrent(path)
	if err != nil {
		return nil, err
	}

	// Printy printy
	fmt.Println("Torrent Info")
	fmt.Println(swarm.Tor.String())

	// Make first contact with tracker
	treq := tracker.NewRequest(swarm.Tor, port, swarm.Id)
	swarm.Tracker, err = tracker.Do(treq)
	if err != nil {
		return nil, err
	}

	// Printy printy
	fmt.Println("\n", swarm.Tracker.String())

	return &swarm, nil
}

func (s *Swarm) Go() error {

	// Start peer Goroutines
	for _, p := range s.Tracker.Peers() {
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
