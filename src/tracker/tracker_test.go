package tracker

import (
	"encoding/binary"
	"fmt"
	"gotor/bencode"
	"net"
	"strings"
	"testing"
)

func TestNewResponse(t *testing.T) {

}

func TestNewResponseCompact(t *testing.T) {

	ips := []net.IP{
		net.ParseIP("255.193.65.23"),
		net.ParseIP("175.178.243.56"),
		net.ParseIP("105.143.135.2"),
		net.ParseIP("147.196.133.77"),
	}

	ports := []uint16{
		30532,
		64005,
		666,
		1337,
	}

	strbuilder := strings.Builder{}
	portBytes := make([]byte, 2)

	strbuilder.WriteString(fmt.Sprintf("%v:", len(ips)*6))
	for i, _ := range ips {
		strbuilder.WriteByte(ips[i][12])
		strbuilder.WriteByte(ips[i][13])
		strbuilder.WriteByte(ips[i][14])
		strbuilder.WriteByte(ips[i][15])
		binary.BigEndian.PutUint16(portBytes, ports[i])
		strbuilder.Write(portBytes)
	}
	strbuilder.WriteByte('e')

	raw := "d8:completei5e10:downloadedi55e10:incompletei2e8:intervali1748e12:min intervali874e5:peers" + strbuilder.String()

	ben, err := bencode.Decode([]byte(raw))
	if err != nil {
		t.Error(err)
	}

	dict, ok := ben.(bencode.Dict)
	if !ok {
		t.Error("failed converting bencode interface to dict")
	}

	resp, err := newResponse(dict)
	if err != nil {
		t.Error(err)
	}

	peers := resp.Peers

	if len(peers) != len(ips) {
		t.Errorf("peer list has length [%v], expected [%v]", len(peers), len(ips))
	}

	// Loop through peers and verify IPs and ports
	for i, p := range peers {
		if !p.Ip().Equal(ips[i]) {
			t.Errorf("bad IP (%v), expected [%v], got [%v]", i, ips[i], p.Ip())
		}
		if p.Port() != ports[i] {
			t.Errorf("bad port (%v), expected [%v], got [%v]", i, ports[i], p.Port())
		}
	}

}
