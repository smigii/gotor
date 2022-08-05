package swarm

import (
	"errors"
	"fmt"
	"log"

	"gotor/p2p"
)

// handleMessage decodes the data in the buffer and handles the messages appropriately.
func (ph *PeerHandler) handleMessage(buf []byte) error {
	var e error
	dar := p2p.DecodeAll(buf)
	pcent := 100.0 * float32(dar.Read) / float32(len(buf))
	log.Printf("Decoded %v/%v (%v%%) bytes from %v\n", dar.Read, len(buf), pcent, ph.peerInfo.Addr())

	for _, msg := range dar.Msgs {
		switch msg.Mtype() {
		case p2p.TypeRequest:
			mreq := msg.(*p2p.MsgRequest)
			e = ph.handleRequest(mreq)
		case p2p.TypeBitfield:
			mbf := msg.(*p2p.MsgBitfield)
			e = ph.handleBitfield(mbf)
		}

		if e != nil {
			return e
		}
	}

	return nil
}

func (ph *PeerHandler) handleBitfield(bfMsg *p2p.MsgBitfield) error {
	// TODO: Check that this is the first message received
	// If not, error and kill connection.
	swarm := ph.swarm

	bf := bfMsg.Bitfield()
	if bf.Nbytes() != swarm.Bf.Nbytes() {
		return errors.New("invalid bitfield")
	}

	// Replace the bitfield
	ph.bf = bf

	// Register
	swarm.PPT.RegisterBF(ph, bf)

	return nil
}

func (ph *PeerHandler) handleRequest(reqMsg *p2p.MsgRequest) error {
	// TODO: Cache pieces
	s := ph.swarm
	idx := int64(reqMsg.Index())
	fmt.Printf("REQUEST SIZE: %v\n", reqMsg.ReqLen())
	if s.Bf.Complete() || s.Bf.Get(idx) {
		_, e := s.Fileio.ReadPiece(idx, ph.buf)
		if e != nil {
			return e
		}
		subdata := ph.buf[reqMsg.Begin() : reqMsg.Begin()+reqMsg.ReqLen()]
		mPiece := p2p.NewMsgPiece(reqMsg.Index(), reqMsg.Begin(), subdata)
		e = s.RLIO.Write(ph.conn, mPiece.Encode())
		if e != nil {
			return e
		}
	}

	return nil
}
