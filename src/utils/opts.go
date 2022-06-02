package utils

import (
	"errors"
	"flag"
	"fmt"
)

// Valid commands
const (
	StartSwarm string = "start-swarm" // Download/Upload
	TorInfo           = "tor-info"    // Read and print torrent info
)

type Opts struct {
	input  *string // Torrent path
	output *string // Output path
	port   *uint   // Listen port
	cmd    *string // What do?
}

// Singleton
var opts *Opts = nil

func GetOpts() *Opts {
	if opts == nil {
		opts = initOpts()
	}
	return opts
}

func initOpts() *Opts {
	opts = &Opts{}
	opts.input = flag.String("i", "", "Path to .torrent file")
	opts.output = flag.String("o", ".", "Output path")
	opts.port = flag.Uint("p", 60666, "Port to listen on")
	opts.cmd = flag.String("cmd", StartSwarm, "Command")

	flag.Parse()

	e := opts.Validate()
	if e != nil {
		panic(e)
	}

	return opts
}

func (o *Opts) Validate() error {
	if o.Input() == "" {
		return errors.New("missing argument 'input'")
	}

	switch *o.cmd {
	case StartSwarm, TorInfo:
		break
	default:
		return fmt.Errorf("invalid command given, [%v]", *o.cmd)
	}
	return nil
}

func (o *Opts) Input() string {
	return *o.input
}

func (o *Opts) Output() string {
	return *o.output
}

func (o *Opts) Port() uint16 {
	return uint16(*o.port)
}

func (o *Opts) Cmd() string {
	return *o.cmd
}