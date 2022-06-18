package swarm

const HandshakeLen = uint8(68)
const HandshakePstrLen = uint8(19)
const HandshakePstr = "BitTorrent protocol"

type Handshake []byte

func MakeHandshake(infohash string, id string) Handshake {
	hs := make(Handshake, HandshakeLen, HandshakeLen)
	hs[0] = HandshakePstrLen
	copy(hs[1:], HandshakePstr)
	copy(hs[28:], infohash)
	copy(hs[48:], id)
	return hs
}

func (hs Handshake) Pstrlen() uint8 {
	return uint8(hs[0])
}

func (hs Handshake) Pstr() []byte {
	return hs[1:20]
}

func (hs Handshake) Reserved() []byte {
	return hs[20:28]
}

func (hs Handshake) Infohash() []byte {
	return hs[28:48]
}

func (hs Handshake) Id() []byte {
	return hs[48:68]
}
