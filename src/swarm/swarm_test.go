package swarm

import (
	"fmt"
	"gotor/torrent"
	"net"
	"testing"
)

func TestSwarm(t *testing.T) {

	//Swarm()

	tor, e := torrent.NewTorrent("../../test_media/medfile.torrent")
	if e != nil {
		panic(e)
	}

	hs := MakeHandshake(tor.Infohash(), "abcdefghijklmnopqrst")
	fmt.Println(hs.Pstrlen())
	fmt.Println(string(hs.Pstr()))
	fmt.Println(hs.Reserved())
	fmt.Println(string(hs.Infohash()))
	fmt.Println(tor.Infohash())
	fmt.Println(string(hs.Id()))

	c, err := net.Dial("tcp", "localhost:9606")
	if err != nil {
		panic(err)
	}

	// 2022-05-30 8:14 P.M. - Detected external IP: 2604:3d09:a580:4b0:f94d:9d3b:7539:91bf

	n, e := c.Write(hs[:])
	if e != nil {
		panic(e)
	}

	fmt.Printf("Sent %v bytes\n", n)

	buf := make([]byte, 4096)
	n, e = c.Read(buf)
	if e != nil {
		panic(e)
	}

	var reply Handshake = buf[:68]

	fmt.Println("REPLY")
	fmt.Println(reply.Pstrlen())
	fmt.Println(string(reply.Pstr()))
	fmt.Println(reply.Reserved())
	fmt.Println(string(reply.Infohash()))
	fmt.Println(tor.Infohash())
	fmt.Println(string(reply.Id()))

	fmt.Printf("Read %v bytes\n", n)
	fmt.Println(string(buf))
}
