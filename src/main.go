package main

import (
	"flag"
	"fmt"
	"gotor/swarm"
	"gotor/utils"
	"log"
)

func main() {

	var in = flag.String("i", "", "Path to .torrent file")
	var out = flag.String("o", ".", "Output path")
	var port = flag.Uint("p", 60666, "Port to listen on")

	flag.Parse()

	opts := utils.GetOpts()
	opts.SetInput(*in)
	opts.SetOutput(*out)
	opts.SetLport(uint16(*port))
	e := opts.Validate()
	if e != nil {
		log.Fatal(e)
	}

	s, e := swarm.NewSwarm(opts)
	if e != nil {
		log.Fatal(e)
	}

	fmt.Println("\n", s.String())

	fmt.Println("\n\nDONE")

}
