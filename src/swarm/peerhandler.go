package swarm

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"gotor/p2p"
	"gotor/peer"
	"gotor/utils/netread"
)

const (
	// NOTE: qBittorrent seems to send a maximum of 8572 bytes per message
	RecvBufSize = 16384

	GetKeepAlive     = 120 * time.Second
	SendKeepAlive    = 60 * time.Second
	HandshakeTimeout = 1 * time.Second
)

// ============================================================================
// STRUCTS ====================================================================

type PeerHandler struct {
	peerInfo  peer.Peer
	peerState peer.State
	swarm     *Swarm
	conn      net.Conn
	chErr     chan<- error // Report errors
	//chReq      <-chan int64    // Pieces we should request from peer
	//chGot      chan<- Got      // Pieces we have successfully written
	procs sync.WaitGroup // How many loops are running for this handler
	buf   []byte         // Buffer for file io operations
}

type Got struct {
	Index int64
	Peer  *PeerHandler
}

// ============================================================================
// FUNK =======================================================================

// Bootstrap creates a TCP connection with the peer, then sends the BitTorrent
// handshake.
func Bootstrap(pInfo peer.Peer, swarm *Swarm) (*PeerHandler, error) {
	conn, e := net.Dial("tcp", pInfo.Addr())

	if e != nil {
		return nil, e
	}

	hs := MakeHandshake(swarm.Tor.Infohash(), swarm.Id)

	_, e = conn.Write(hs)
	if e != nil {
		return nil, e
	}

	torInfo := swarm.Tor.Info()

	return &PeerHandler{
		peerInfo: pInfo,
		conn:     conn,
		swarm:    swarm,
		chErr:    swarm.ChErr,
		procs:    sync.WaitGroup{},
		buf:      make([]byte, torInfo.PieceLen(), torInfo.PieceLen()),
	}, nil
}

// Incoming receives a new peer connection. It will first check for the correct
// BitTorrent handshake, add to the peer list, then send a handshake and bitfield back.
func Incoming(c net.Conn, swarm *Swarm) (*PeerHandler, error) {

	// Must be using TCP (for now atleast)
	tcpAddr, ok := c.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("connection is not TCP")
	}

	buf := make([]byte, HandshakeLen)

	// Set timeout
	e := c.SetReadDeadline(time.Now().Add(HandshakeTimeout))
	if e != nil {
		return nil, e
	}

	// Read the handshake
	_, e = c.Read(buf)
	if e != nil {
		return nil, e
	}
	peerHs := Handshake(buf)
	if !Validate(peerHs, swarm.Tor.Infohash()) {
		_ = c.Close() // TODO: Handle?
		return nil, fmt.Errorf("bad peer handshake")
	}
	log.Printf("good handshake from %v", c.RemoteAddr())

	// Send handshake
	hs := MakeHandshake(swarm.Tor.Infohash(), swarm.Id)
	_, e = c.Write(hs)
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v handshake\n", c.RemoteAddr())

	// Send bitfield
	bfmsg := p2p.NewMsgBitfield(swarm.Bf)
	_, e = c.Write(bfmsg.Encode())
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v bitfield\n", c.RemoteAddr())

	newPeer := peer.MakePeer(string(peerHs.Id()), tcpAddr.IP, uint16(tcpAddr.Port))

	torInfo := swarm.Tor.Info()

	return &PeerHandler{
		peerInfo: newPeer,
		swarm:    swarm,
		conn:     c,
		chErr:    swarm.ChErr,
		procs:    sync.WaitGroup{},
		buf:      make([]byte, torInfo.PieceLen(), torInfo.PieceLen()),
	}, nil
}

// ============================================================================
// ============================================================================

func (ph *PeerHandler) Loop() {

	// Any goroutines spawned should report errors on this chan
	chErr := make(chan error)

	// Use to stop to all spawned goroutines
	chDone := make(chan bool)

	// For now, just unchoke everyone
	msg := p2p.NewMsgUnchoke()
	_, e := ph.conn.Write(msg.Encode())
	if e != nil {
		log.Printf("error unchoking: %v\n", e)
	}

	go ph.pingLoop(chErr, chDone)

	go ph.recvLoop(chErr, chDone)

	done := false
	for {
		if done {
			break
		}

		select {
		// For now, just kill ourselves if we receive any error.
		// We will fine-tune this later
		case e = <-chErr:
			done = true
			log.Printf("error peer %v (killing): %v", ph.peerInfo.Addr(), e)
			chDone <- true
			ph.procs.Wait()
			// We will eventually wrap this in a struct so that we can
			// tell the main loop which PeerHandler has errored
			ph.chErr <- e
		}
	}

	log.Printf("peer %v done", ph.peerInfo.Addr())
}

// recvLoop handles reading in data from the peer and handling replies.
// To terminate the loop, call cancelRecvLoop.
func (ph *PeerHandler) recvLoop(chErr chan<- error, chKill <-chan bool) {
	// TODO: Should probably handle split messages
	// What if we need to recv more than we can handle?
	// 1. Decode all messages
	// 2. Any subsections of bytes that were not decoded AND come
	//    before any successfully decoded messages are discarded
	// 3. This could leave a tail of non-decoded bytes. These should
	//    be stored in another buffer. When we go to read again,
	//    append new data to previous data and try decoding all again.
	// 4. Rinse repeat.

	ph.procs.Add(1)
	defer ph.procs.Done()

	defer log.Printf("end recvLoop %v", ph.peerInfo.Addr())
	log.Printf("start recvLoop %v", ph.peerInfo.Addr())

	readLoop := netread.NewReadLoop(RecvBufSize, ph.conn, GetKeepAlive)
	go readLoop.Run()

	var e error
	done := false
	for !done {

		select {
		case buf := <-readLoop.ReadChBuf():
			e = ph.handleMessage(buf)
			readLoop.Ready()
		case e = <-readLoop.ReadChErr():
			done = true
		case <-chKill:
			readLoop.Finish()
			done = true
		}

		if e != nil {
			chErr <- e
			done = true
		}
	}
}

// pingLoop sends a keep alive message to the peer at a set interval.
func (ph *PeerHandler) pingLoop(chErr chan<- error, chDone <-chan bool) {

	ph.procs.Add(1)
	defer ph.procs.Done()
	defer log.Printf("end pingLoop %v", ph.peerInfo.Addr())

	log.Printf("start pingLoop %v", ph.peerInfo.Addr())

	addr := ph.peerInfo.Addr()
	ticker := time.NewTicker(SendKeepAlive)
	ka := p2p.KeepAliveSingleton
	data := ka.Encode()

	done := false
	for !done {
		select {
		case <-ticker.C:
			_, e := ph.conn.Write(data)
			if e != nil {
				chErr <- e
				log.Printf("error with keep alive to %v", addr)
				done = true
			}
			log.Printf("sent keep alive to %v", addr)
		case <-chDone:
			done = true
		}
	}
}
