package peer

import (
	"fmt"
	"net"
	"strings"
)

type Peer struct {
	id           string
	ip           net.IP
	port         uint16
	addr         string
	chokingUs    bool // Peer is choking us
	weChoking    bool // We are choking peer
	interestedUs bool // Peer is interested in us
	weInterested bool // We are interested in peer
}

func NewPeer(id string, ip net.IP, port uint16) *Peer {
	strb := strings.Builder{}
	strb.WriteString(ip.String())
	strb.WriteByte(':')
	strb.WriteString(fmt.Sprintf("%v", port))

	return &Peer{
		id:           id,
		ip:           ip,
		port:         port,
		addr:         strb.String(),
		chokingUs:    true,
		weChoking:    true,
		interestedUs: false,
		weInterested: false,
	}
}

func (p Peer) Id() string {
	return p.id
}

func (p Peer) Ip() net.IP {
	return p.ip
}

func (p Peer) Port() uint16 {
	return p.port
}

func (p *Peer) Addr() string {
	return p.addr
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
