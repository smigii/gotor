package utils

import "errors"

type Opts struct {
	input  string // Torrent path
	output string // Output path
	lport  uint16 // Listen port
}

// Singleton
var opts *Opts = nil

func GetOpts() *Opts {
	if opts == nil {
		opts = &Opts{}
	}
	return opts
}

func (o *Opts) Validate() error {
	if o.Input() == "" {
		return errors.New("missing argument 'input'")
	}
	return nil
}

func (o *Opts) Input() string {
	return o.input
}

func (o *Opts) SetInput(input string) {
	o.input = input
}

func (o *Opts) Output() string {
	return o.output
}

func (o *Opts) SetOutput(output string) {
	o.output = output
}

func (o *Opts) Lport() uint16 {
	return o.lport
}

func (o *Opts) SetLport(lport uint16) {
	o.lport = lport
}
