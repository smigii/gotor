package swarm

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"gotor/bf"
	"gotor/p2p"
	"gotor/peer"
	"gotor/torrent"
	"gotor/torrent/fileio"
	"gotor/tracker"
	"gotor/utils"
)

const (
	KeepAliveTimer   = 60 * time.Second
	KeepAliveTimeout = 120 * time.Second
)

// ============================================================================
// STRUCTS ====================================================================

type Swarm struct {
	State  *tracker.State
	Stats  *tracker.Stats
	Peers  peer.List
	Tor    *torrent.Torrent
	Fileio *fileio.FileIO
	Bf     *bf.Bitfield
	Id     string
	Port   uint16
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
	swarm.Tor, err = torrent.FromTorrentFile(opts.Input(), opts.WorkingDir())
	if err != nil {
		return nil, err
	}

	torInfo := swarm.Tor.Info()

	// Make the FileIO handler
	swarm.Fileio = fileio.NewFileIO()

	// OCAT files
	log.Printf("openning and validating files")
	e := swarm.Fileio.OCATAll(torInfo.Files())
	if e != nil {
		return nil, e
	}

	// Make bitfield
	swarm.Bf = bf.NewBitfield(torInfo.NumPieces())
	e = swarm.Validate()
	if e != nil {
		return nil, e
	}
	_bf := swarm.Bf
	pcent := 100 * float64(_bf.Nset()) / float64(_bf.Nbits())
	log.Printf("have %v/%v (%v%%) pieces", _bf.Nset(), _bf.Nbits(), pcent)

	// TODO: Compute remaining bytes left
	//swarm.Stats = tracker.NewStats(0, 0, swarm.Tor.Length())  // Full leech
	swarm.Stats = tracker.NewStats(0, 0, 0) // Seed

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

func (s *Swarm) Validate() error {

	torInfo := s.Tor.Info()
	buf := make([]byte, torInfo.PieceLen(), torInfo.PieceLen())

	var i int64
	for i = 0; i < torInfo.NumPieces(); i++ {

		n, e := s.Fileio.ReadPiece(i, torInfo, buf)
		if e != nil {
			return e
		}

		knownHash := torInfo.PieceHash(i)
		gotHash := utils.SHA1(buf[:n])
		eq := knownHash == gotHash
		s.Bf.Set(i, eq)
	}

	return nil
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
	bfmsg := p2p.NewMsgBitfield(s.Bf)
	_, e = c.Write(bfmsg.Encode())
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v bitfield\n", c.RemoteAddr())

	newPeer := peer.NewPeer(string(peerHs.Id()), tcpAddr.IP, uint16(tcpAddr.Port))
	newPeer.Conn = c
	return newPeer, nil
}

func pingLoop(peer *peer.Peer) {
	ticker := time.NewTicker(KeepAliveTimer)
	ka := p2p.KeepAliveSingleton
	data := ka.Encode()
	for {
		select {
		case <-ticker.C:
			_, e := peer.Conn.Write(data)
			if e != nil {
				panic(e)
			}
			log.Printf("sent keep alive to %v", peer.Id())
		}
	}
}

func (s *Swarm) HandlePeer(peer *peer.Peer) {

	go pingLoop(peer)

	// For now, just unchoke everyone
	msg := p2p.NewMsgUnchoke()
	_, e := peer.Conn.Write(msg.Encode())
	if e != nil {
		log.Printf("error unchoking: %v\n", e)
	}

	recvBuf := make([]byte, 4096)

	for {
		n, e := peer.Conn.Read(recvBuf)
		if n == 0 || e != nil {
			log.Printf("Peer dead %v", peer.Conn.RemoteAddr())
			break
		}

		fmt.Printf("\n--- MESSAGE ---\n")
		fmt.Println(recvBuf[:n])
		fmt.Printf("--- END ---\n\n")

		dar := p2p.DecodeAll(recvBuf[:n])
		pcent := 100.0 * float32(dar.Read) / float32(n)
		log.Printf("Decoded %v/%v (%v%%) bytes from %v\n\n", dar.Read, n, pcent, peer.Conn.RemoteAddr())
		for _, msg := range dar.Msgs {
			fmt.Printf("%v\n\n", msg.String())
			switch msg.Mtype() {
			case p2p.TypeRequest:
				mreq := msg.(*p2p.MsgRequest)
				e := s.handleRequest(peer, mreq)
				if e != nil {
					panic(e)
				}
			}
		}
	}
}

func (s *Swarm) handleRequest(peer *peer.Peer, req *p2p.MsgRequest) error {
	// TODO: This is so inefficient it hurts
	torInfo := s.Tor.Info()
	idx := int64(req.Index())
	if s.Bf.Complete() || s.Bf.Get(idx) {
		buf := make([]byte, torInfo.PieceLen(), torInfo.PieceLen())
		_, e := s.Fileio.ReadPiece(idx, s.Tor.Info(), buf)
		if e != nil {
			return e
		}
		subdata := buf[req.Begin() : req.Begin()+req.ReqLen()]
		mPiece := p2p.NewMsgPiece(req.Index(), req.Begin(), subdata)
		_, e = peer.Conn.Write(mPiece.Encode())
		if e != nil {
			return e
		}
	}

	return nil
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
