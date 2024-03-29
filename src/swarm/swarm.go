package swarm

import (
	"fmt"
	"log"
	"net"
	"strings"

	"gotor/bf"
	"gotor/io"
	"gotor/peer"
	"gotor/torrent"
	"gotor/torrent/fileio"
	"gotor/tracker"
	"gotor/utils"
)

// ============================================================================
// STRUCTS ====================================================================

type Swarm struct {
	State  *tracker.State
	Stats  *tracker.Stats
	Peers  peer.List
	Tor    *torrent.Torrent
	Fileio *fileio.FileIO
	RLIO   *io.RateLimitIO
	Bf     *bf.Bitfield
	PPT    *PeerPieceTracker
	Id     string
	Port   uint16

	ChErr chan error
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
	swarm.Fileio = fileio.NewFileIO(torInfo)

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

	swarm.RLIO = io.NewRateLimitIO()
	swarm.RLIO.SetWriteRate(opts.UpLimit())
	swarm.RLIO.SetReadRate(opts.DnLimit())

	swarm.PPT = NewPeerPieceTracker(uint32(torInfo.NumPieces()), swarm.Bf)

	return &swarm, nil
}

func (s *Swarm) Validate() error {

	torInfo := s.Tor.Info()
	buf := make([]byte, torInfo.PieceLen(), torInfo.PieceLen())

	var i int64
	for i = 0; i < torInfo.NumPieces(); i++ {

		n, e := s.Fileio.ReadPiece(i, buf)
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

	go s.runListener()
	go s.RLIO.Run()

	// Start peer Goroutines
	for _, p := range s.Peers {
		go func(peer peer.Info) {
			ph, e := FromBootstrap(peer, s)
			if e != nil {
				log.Printf("failed to bootstrap %v : %v", peer.Addr(), e)
			} else {
				ph.Loop()
			}
		}(p)
	}
}

func (s *Swarm) runListener() {
	opts := utils.GetOpts()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", opts.Port()))
	if err != nil {
		panic(err)
	}
	log.Printf("Listening on port %v\n", opts.Port())

	phs := make([]*PeerHandler, 0, 4)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		log.Printf("new client @ %v", conn.RemoteAddr())

		go func(c net.Conn) {
			ph, e := FromIncoming(c, s)
			if e != nil {
				log.Println(e)
			} else {
				phs = append(phs, ph)
				ph.Loop()
			}
		}(conn)
	}
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
