package main

import (
	"fmt"
	"gotor/swarm"
	"gotor/torrent"
	"gotor/utils"
	"log"
)

func main() {

	opts := utils.GetOpts()

	switch opts.Cmd() {
	case utils.StartSwarm:
		CmdSwarm(opts)
	case utils.TorInfo:
		CmdTorInfo(opts)
	default:
		fmt.Printf("invalid command [%v]", opts.Cmd())
	}

	fmt.Println("\n\nDONE")

}

func CmdSwarm(opts *utils.Opts) {
	s, e := swarm.NewSwarm(opts)
	if e != nil {
		log.Fatal(e)
	}

	fmt.Println("\n", s.String())

	s.Start()

	for {

	}
}

func CmdTorInfo(opts *utils.Opts) {
	tor, e := torrent.NewTorrent(opts.Input())
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println(tor.String())
}
