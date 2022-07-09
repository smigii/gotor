package swarm

import (
	"gotor/p2p"
)

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
