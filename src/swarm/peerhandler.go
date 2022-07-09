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
)

type PeerHandler struct {
	peerInfo  peer.Peer
	peerState peer.State
	conn      net.Conn
	swarm     *Swarm
	wg        sync.WaitGroup
}

const (
	GetKeepAlive  = 120 * time.Second
	SendKeepAlive = 60 * time.Second
)

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

	return &PeerHandler{
		peerInfo: pInfo,
		conn:     conn,
		swarm:    swarm,
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

	// TODO: Set timeout
	// Read the handshake
	_, e := c.Read(buf)
	if e != nil {
		return nil, e
	}
	peerHs := Handshake(buf)
	if !Validate(peerHs, swarm.Tor.Infohash()) {
		_ = c.Close() // TODO: Handle?
		return nil, fmt.Errorf("bad peer handshake")
	}
	log.Printf("good handshake @ %v", c.RemoteAddr())

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
	return &PeerHandler{
		peerInfo: newPeer,
		conn:     c,
		swarm:    swarm,
	}, nil
}

// ============================================================================
// ============================================================================

func (ph *PeerHandler) Start() {

	// Any goroutines spawned should report errors on this chan
	errChan := make(chan error)

	// Use to stop to all spawned goroutines
	killChan := make(chan bool)

	// For now, just unchoke everyone
	msg := p2p.NewMsgUnchoke()
	_, e := ph.conn.Write(msg.Encode())
	if e != nil {
		log.Printf("error unchoking: %v\n", e)
	}

	go ph.pingLoop(errChan, killChan)

	go ph.recvLoop(errChan)

	done := false
	for {
		if done {
			break
		}

		select {
		// For now, just kill ourselves if we receive any error.
		// We will fine-tune this later
		case e = <-errChan:
			done = true
			log.Printf("error peer %v (killing): %v", ph.peerInfo.Addr(), e)
			killChan <- true
			_ = ph.cancelRecvLoop()
			ph.wg.Wait()
			// We will eventually wrap this in a struct so that we can
			// tell the main loop which PeerHandler has errored
			ph.swarm.ChErr <- e
		}
	}

	log.Printf("peer %v done", ph.peerInfo.Addr())
}

// recvLoop handles reading in data from the peer and handling replies.
// To terminate the loop, call cancelRecvLoop. TODO: I don't like this
func (ph *PeerHandler) recvLoop(errChan chan<- error) {
	// TODO: Should probably handle split messages
	// What if we need to recv more than we can handle?
	// 1. Decode all messages
	// 2. Any subsections of bytes that were not decoded AND come
	//    before any successfully decoded messages are discarded
	// 3. This could leave a tail of non-decoded bytes. These should
	//    be stored in another buffer. When we go to read again,
	//    append new data to previous data and try decoding all again.
	// 4. Rinse repeat.

	ph.wg.Add(1)
	defer ph.wg.Done()
	defer log.Printf("end recvLoop %v", ph.peerInfo.Addr())

	log.Printf("start recvLoop %v", ph.peerInfo.Addr())

	// NOTE: qBittorrent seems to send a maximum of 8572 bytes per message
	recvBuf := make([]byte, 16384)

	for {

		// Check for data. If we don't hear from them within GetKeepAlive time,
		// assume peer has suffered a tragic fate.
		e := ph.conn.SetReadDeadline(time.Now().Add(GetKeepAlive))
		if e != nil {
			errChan <- e
			break
		}

		n, e := ph.conn.Read(recvBuf)
		if n == 0 || e != nil {
			errChan <- e
			break
		}

		dar := p2p.DecodeAll(recvBuf[:n])
		pcent := 100.0 * float32(dar.Read) / float32(n)
		log.Printf("Decoded %v/%v (%v%%) bytes from %v\n", dar.Read, n, pcent, ph.peerInfo.Addr())
		for _, msg := range dar.Msgs {
			//fmt.Printf("%v\n\n", msg.String())
			switch msg.Mtype() {
			case p2p.TypeRequest:
				mreq := msg.(*p2p.MsgRequest)
				e = ph.handleRequest(mreq)
			}

			if e != nil {
				errChan <- e
				break
			}
		}
	}
}

// pingLoop sends a keep alive message to the peer at a set interval.
func (ph *PeerHandler) pingLoop(errChan chan<- error, killChan <-chan bool) {

	ph.wg.Add(1)
	defer ph.wg.Done()
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
				errChan <- e
				log.Printf("error with keep alive to %v", addr)
				done = true
			}
			log.Printf("sent keep alive to %v", addr)
		case <-killChan:
			done = true
		}
	}
}

func (ph *PeerHandler) cancelRecvLoop() error {
	// Set the cancel time to now. Rather than using time.Now() which involves
	// a syscall, use Unix(1,0)
	return ph.conn.SetReadDeadline(time.Unix(1, 0))
}
