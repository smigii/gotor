package peer

import (
	"fmt"
	"net"
	"strings"
)

type Peer struct {
	id   string
	ip   net.IP
	port uint16
	addr string
}

func MakePeer(id string, ip net.IP, port uint16) Peer {
	strb := strings.Builder{}
	strb.WriteString(ip.String())
	strb.WriteByte(':')
	strb.WriteString(fmt.Sprintf("%v", port))

	return Peer{
		id:   id,
		ip:   ip,
		port: port,
		addr: strb.String(),
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
