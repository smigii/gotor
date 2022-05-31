package swarm

import (
	"gotor/torrent"
	"gotor/tracker"
)

const (
	Host = "localhost"
	Port = "60666"
)

type Swarm struct {
	torrent *torrent.Torrent
	tracker *tracker.Resp
}

//func bootstrap(peer peer.Peer) error {
//var err error
//peer.conn, err = net.Dial("tcp", peer.Addr())
//
//if err != nil {
//	return err
//}
//
//hs := MakeHandshake()
//}
