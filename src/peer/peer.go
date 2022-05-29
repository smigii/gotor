package peer

type Peer struct {
	id           string
	ip           string
	port         uint16
	chokingUs    bool // Peer is choking us
	weChoking    bool // We are choking peer
	interestedUs bool // Peer is interested in us
	weInterested bool // We are interested in peer
}

func NewPeer(id string, ip string, port uint16) *Peer {
	return &Peer{
		id:           id,
		ip:           ip,
		port:         port,
		chokingUs:    true,
		weChoking:    true,
		interestedUs: false,
		weInterested: false,
	}
}

func (p Peer) Id() string {
	return p.id
}

func (p Peer) Ip() string {
	return p.ip
}

func (p Peer) Port() uint16 {
	return p.port
}

func (p *Peer) ChokingUs() bool {
	return p.chokingUs
}

func (p *Peer) SetChokingUs(chokingUs bool) {
	p.chokingUs = chokingUs
}

func (p *Peer) WeChoking() bool {
	return p.weChoking
}

func (p *Peer) SetWeChoking(weChoking bool) {
	p.weChoking = weChoking
}

func (p *Peer) InterestedUs() bool {
	return p.interestedUs
}

func (p *Peer) SetInterestedUs(interestedUs bool) {
	p.interestedUs = interestedUs
}

func (p *Peer) WeInterested() bool {
	return p.weInterested
}

func (p *Peer) SetWeInterested(weInterested bool) {
	p.weInterested = weInterested
}
