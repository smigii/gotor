package peer

import (
	"fmt"
	"net"
	"strings"
)

type Info struct {
	id   string
	ip   net.IP
	port uint16
	addr string
	str  string
}

func MakePeer(id string, ip net.IP, port uint16) Info {
	strb := strings.Builder{}
	strb.WriteString(ip.String())
	strb.WriteByte(':')
	strb.WriteString(fmt.Sprintf("%v", port))

	return Info{
		id:   id,
		ip:   ip,
		port: port,
		addr: strb.String(),
		str:  fmt.Sprintf("%v @ %v:%v", id, ip, port),
	}
}

func (p Info) Id() string {
	return p.id
}

func (p Info) Ip() net.IP {
	return p.ip
}

func (p Info) Port() uint16 {
	return p.port
}

func (p *Info) Addr() string {
	return p.addr
}

func (p *Info) String() string {
	return p.str
}
