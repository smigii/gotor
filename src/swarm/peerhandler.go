package swarm

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"gotor/bf"
	"gotor/io"
	"gotor/p2p"
	"gotor/peer"
)

const (
	// NOTE: qBittorrent seems to send a maximum of 8572 bytes per message
	RecvBufSize = 16384

	GetKeepAlive     = 120 * time.Second
	SendKeepAlive    = 60 * time.Second
	HandshakeTimeout = 1 * time.Second

	// requestLength is the same piece request length as qBittorrent
	requestLength = 16384
)

// ============================================================================
// STRUCTS ====================================================================

type PeerHandler struct {
	peerInfo  peer.Info
	peerState peer.State
	swarm     *Swarm
	conn      net.Conn
	bf        *bf.Bitfield
	procs     sync.WaitGroup // How many loops are running for this handler
	buf       []byte         // Buffer for file io operations

	chErr chan<- error // Report errors
}

// ============================================================================
// FUNK =======================================================================

// FromBootstrap creates a TCP connection with the peer, then sends the BitTorrent
// handshake.
func FromBootstrap(pInfo peer.Info, swarm *Swarm) (*PeerHandler, error) {
	conn, e := net.Dial("tcp", pInfo.Addr())
	if e != nil {
		return nil, e
	}

	hs := MakeHandshake(swarm.Tor.Infohash(), swarm.Id)

	_, e = conn.Write(hs)
	if e != nil {
		return nil, e
	}

	return NewPeerHandler(pInfo, swarm, conn), nil
}

// FromIncoming receives a new peer connection. It will first check for the correct
// BitTorrent handshake, add to the peer list, then send a handshake and bitfield back.
func FromIncoming(conn net.Conn, swarm *Swarm) (*PeerHandler, error) {

	// Must be using TCP (for now atleast)
	tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("connection is not TCP")
	}

	buf := make([]byte, HandshakeLen)

	// Set timeout
	e := conn.SetReadDeadline(time.Now().Add(HandshakeTimeout))
	if e != nil {
		return nil, e
	}

	// Read the handshake
	_, e = conn.Read(buf)
	if e != nil {
		return nil, e
	}
	peerHs := Handshake(buf)
	if !ValidHandshake(peerHs, swarm.Tor.Infohash()) {
		_ = conn.Close() // TODO: Handle?
		return nil, fmt.Errorf("bad peer handshake")
	}
	log.Printf("good handshake from %v", conn.RemoteAddr())

	// Send handshake
	hs := MakeHandshake(swarm.Tor.Infohash(), swarm.Id)
	_, e = conn.Write(hs)
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v handshake\n", conn.RemoteAddr())

	// Send bitfield
	bfmsg := p2p.NewMsgBitfield(swarm.Bf)
	_, e = conn.Write(bfmsg.Encode())
	if e != nil {
		return nil, e
	}
	log.Printf("Sent %v bitfield\n", conn.RemoteAddr())

	newPeer := peer.MakePeer(string(peerHs.Id()), tcpAddr.IP, uint16(tcpAddr.Port))

	return NewPeerHandler(newPeer, swarm, conn), nil
}

func NewPeerHandler(pInfo peer.Info, swarm *Swarm, conn net.Conn) *PeerHandler {
	torInfo := swarm.Tor.Info()
	return &PeerHandler{
		peerInfo: pInfo,
		swarm:    swarm,
		conn:     conn,
		chErr:    swarm.ChErr,
		procs:    sync.WaitGroup{},
		buf:      make([]byte, torInfo.PieceLen(), torInfo.PieceLen()),
		bf:       bf.NewBitfield(torInfo.NumPieces()),
	}
}

// ============================================================================
// ============================================================================

func (ph *PeerHandler) Choke() error {
	msg := p2p.NewMsgUnchoke()
	_, e := ph.conn.Write(msg.Encode())
	return e
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

	go ph.requestLoop(chErr, chDone)

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
			log.Printf("error peer [%v] (killing): %v", ph.peerInfo.String(), e)
			chDone <- true
			ph.procs.Wait()
			// We will eventually wrap this in a struct so that we can
			// tell the main loop which PeerHandler has errored
			ph.chErr <- e
		}
	}

	log.Printf("peer %v done", ph.peerInfo.Addr())
}

// recvLoop handles reading in data from the peer and sending
// replies if needed.
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

	defer log.Printf("end recvLoop [%v]", ph.peerInfo.String())
	log.Printf("start recvLoop [%v]", ph.peerInfo.String())

	readLoop := io.NewReadLoop(RecvBufSize, ph.conn, GetKeepAlive)
	go readLoop.Run()

	var e error
	done := false
	for !done {

		select {
		case buf := <-readLoop.ReadData():
			e = ph.handleMessage(buf)
			readLoop.Ready()
		case e = <-readLoop.ReadError():
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

// requestLoop sends out piece requests to the peer.
func (ph *PeerHandler) requestLoop(chErr chan<- error, chDone <-chan bool) {
	ph.procs.Add(1)
	defer ph.procs.Done()

	log.Printf("start requestLoop [%v]", ph.peerInfo.String())
	defer log.Printf("end requestLoop [%v]", ph.peerInfo.String())

	reqs := make([]uint32, 0, 5)

	// TODO: Yikes
	for !ph.swarm.Bf.Complete() {

		// Fill up requests
		for i := len(reqs); i < 5; i++ {
			next, ok := ph.swarm.PPT.NextPiece(ph)
			// TODO: Handle this
			if !ok {
				break
			}
			reqs = append(reqs, next)

			// Make the request. qBittorrent uses 16KiB requests
			msgs := ph.createReqMessages(next)
			for _, msg := range msgs {
				// TODO: Handle
				_ = ph.swarm.RLIO.Write(ph.conn, msg.Encode())
				fmt.Printf("sent request for %v : %v", msg.Index(), msg.ReqLen())
			}
			fmt.Printf("sent %v msgs", len(msgs))
		}

	}
}

func (ph *PeerHandler) createReqMessages(index uint32) []*p2p.MsgRequest {
	msgs := make([]*p2p.MsgRequest, 0, 3)
	ti := ph.swarm.Tor.Info()

	length := uint32(ti.PieceLen())
	if int64(index) == ti.NumPieces()-1 {
		length = uint32(ti.LastPieceLen())
	}

	offset := uint32(0)

	for offset < length {
		reqlen := uint32(math.Min(requestLength, float64(length-offset)))
		msgs = append(msgs, p2p.NewMsgRequest(index, offset, reqlen))
	}

	return msgs
}

// pingLoop sends a keep alive message to the peer at a set interval
// defined by SendKeepAlive.
func (ph *PeerHandler) pingLoop(chErr chan<- error, chDone <-chan bool) {

	ph.procs.Add(1)
	defer ph.procs.Done()
	defer log.Printf("end pingLoop [%v]", ph.peerInfo.String())

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

func (ph *PeerHandler) Key() string {
	return ph.peerInfo.String()
}
