package swarm

import (
	"errors"
	"fmt"
	"gotor/p2p"
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
	State    *tracker.State
	Stats    *tracker.Stats
	Peers    peer.List
	Tor      *torrent.Torrent
	Id       string
	Port     uint16
	Bitfield utils.Bitfield
}

// ============================================================================
// FUNK =======================================================================

func NewSwarm(opts *utils.Opts) (*Swarm, error) {
	var err error

	swarm := Swarm{}
	swarm.Id = utils.NewPeerId()
	swarm.Port = opts.Port()

	// Read torrent file
	log.Printf("reading torrent file [%v]\n", opts.Input())
	swarm.Tor, err = torrent.NewTorrent(opts.Input())
	if err != nil {
		return nil, err
	}

	// TODO: Check opts.output and see how much is really done
	// For now, we're just acting as a server, so fill bitfield
	//swarm.Stats = tracker.NewStats(0, 0, swarm.Tor.Length())
	swarm.Stats = tracker.NewStats(0, 0, 0)
	swarm.Bitfield = utils.NewBitfield(swarm.Tor.NumPieces())
	swarm.Bitfield.Fill()

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

func (s *Swarm) Start() {

	go s.listen()

	// Start peer Goroutines
	for _, p := range s.Peers {
		go func(peer *peer.Peer) {
			e := s.bootstrap(peer)
			if e != nil {
				// TODO: KILL THE PEER
			} else {
				s.HandlePeer(peer)
			}
		}(p)
	}

}

func (s *Swarm) listen() {
	opts := utils.GetOpts()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", opts.Port()))
	if err != nil {
		panic(err)
	}
	log.Printf("Listening on port %v\n", opts.Port())

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		log.Printf("New client @ %v", conn.RemoteAddr())
		go func() {
			p, e := s.incomingPeer(conn)
			if e != nil {
				log.Println(e)
			} else {
				s.HandlePeer(p)
			}
		}()
	}
}

// incomingPeer receives a new peer connection. It will first check for the correct
// BitTorrent handshake, add to the peer list, then send a handshake and bitfield back.
func (s *Swarm) incomingPeer(c net.Conn) (*peer.Peer, error) {

	// Must be using TCP (for now atleast)
	tcpAddr, ok := c.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("connection is not TCP")
	}

	buf := make([]byte, HandshakeLen)

	// Read the handshake
	n, e := c.Read(buf)
	if e != nil {
		return nil, e
	}
	peerHs := Handshake(buf)
	// TODO: Check more
	if n != int(HandshakeLen) || string(peerHs.Infohash()) != s.Tor.Infohash() {
		//c.Close()
		//return nil, errors.New("bad peer handshake")
	}

	// Send handshake
	hs := MakeHandshake(s.Tor.Infohash(), s.Id)
	_, e = c.Write(hs)
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v handshake\n", c.RemoteAddr())

	// Send bitfield
	msg := p2p.NewMsgBitfield(s.Bitfield.Data())
	_, e = c.Write(msg.Encode())
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v bitfield\n", c.RemoteAddr())

	newPeer := peer.NewPeer(string(peerHs.Id()), tcpAddr.IP, uint16(tcpAddr.Port))
	newPeer.Conn = c
	return newPeer, nil
}

func (s *Swarm) HandlePeer(peer *peer.Peer) {

	// For now, just unchoke everyone
	msg := p2p.NewMsgUnchoke()
	_, e := peer.Conn.Write(msg.Encode())
	if e != nil {
		log.Printf("error unchoking: %v\n", e)
	}

	buf := make([]byte, 4096)
	for {
		n, e := peer.Conn.Read(buf)
		if n == 0 || e != nil {
			log.Printf("Peer dead %v", peer.Conn.RemoteAddr())
			break
		}

		fmt.Println("--- MESSAGE ---")
		fmt.Println(buf[:n])
		fmt.Println("--- END ---")

		msg, e := p2p.Decode(buf)
		if e != nil {
			log.Println(e)
		} else {
			log.Printf("New message from %v\n", peer.Conn.RemoteAddr())
			fmt.Println(msg.String())
		}
	}
}

// bootstrap
// Create a TCP connection with the peer, then send
// the BitTorrent handshake.
func (s *Swarm) bootstrap(peer *peer.Peer) error {
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
