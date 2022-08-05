package swarm

import (
	"errors"
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
	swarm := ph.swarm
	torInfo := swarm.Tor.Info()

	bf := bfMsg.Bitfield()
	if bf.Nbytes() != swarm.Bf.Nbytes() {
		return errors.New("invalid bitfield")
	}

	// Set peer bitfield. Since we aren't encoding this value ever, we can
	// just store it as a simple slice of bools
	havePieces := make([]int64, 0, torInfo.NumPieces())
	for i := int64(0); i < torInfo.NumPieces(); i++ {
		val := bf.Get(i)
		if val {
			ph.pieces[i] = true
			havePieces = append(havePieces, i)
		}
	}

	// Increment all the pieces we have
	swarm.Ppt.Register(ph, havePieces...)

	return nil
}

func (ph *PeerHandler) handleRequest(reqMsg *p2p.MsgRequest) error {
	// TODO: Cache pieces
	s := ph.swarm
	idx := int64(reqMsg.Index())
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
