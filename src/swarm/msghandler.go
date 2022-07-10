package swarm

import (
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
		}

		if e != nil {
			return e
		}
	}

	return nil
}

func (ph *PeerHandler) handleRequest(req *p2p.MsgRequest) error {
	// TODO: This is so inefficient it hurts

	s := ph.swarm
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
		_, e = ph.conn.Write(mPiece.Encode())
		if e != nil {
			return e
		}
	}

	return nil
}
