package utils

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// Valid commands
const (
	StartSwarm string = "swarm" // Download/Upload
	TorInfo           = "info"  // Read and print torrent info
)

type Opts struct {
	input *string // Input torrent path
	wd    *string // Working directory path
	port  *uint   // Listen port
	cmd   *string // What do?

	uplimStr *string
	dnlimStr *string
	uplim    int64 // Upload limit in bytes / sec
	dnlim    int64 // Download limit in bytes / sec
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
	opts.wd = flag.String("w", "", "Working directory")
	opts.port = flag.Uint("p", 60666, "Port to listen on")
	opts.cmd = flag.String("cmd", StartSwarm, "Command")

	opts.uplimStr = flag.String("u", "-1B", "Upload limit in form X[B|K|M|G]")
	opts.dnlimStr = flag.String("d", "-1B", "Download limit in form X[B|K|M|G]")

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

	if o.Input() == "" {
		return errors.New("missing argument 'working directory'")
	}

	switch *o.cmd {
	case StartSwarm, TorInfo:
		break
	default:
		return fmt.Errorf("invalid command given, [%v]", *o.cmd)
	}

	// Upload limit
	v, e := parseSizeUnits(*opts.uplimStr)
	if e != nil {
		return e
	} else {
		opts.uplim = v
	}

	// Download limit
	v, e = parseSizeUnits(*opts.dnlimStr)
	if e != nil {
		return e
	} else {
		opts.dnlim = v
	}

	return nil
}

func parseSizeUnits(str string) (int64, error) {
	str = strings.ToUpper(str)
	lenstr := len(str)
	if lenstr < 1 {
		return 0, errors.New("empty string")
	}

	i := 0
	mult := int64(1)
	if str[0] == '-' {
		mult = -1
		i++
	}

	strb := strings.Builder{}

	// Read the numbers
	for ; i < lenstr; i++ {
		if str[i] >= '0' && str[i] <= '9' {
			strb.WriteByte(str[i])
		} else {
			break
		}
	}

	v, e := strconv.Atoi(strb.String())
	if e != nil {
		return 0, e
	}
	val := int64(v) * mult

	// No letter, assume units BYTES
	if i == lenstr {
		return val, nil
	}

	// If there are still more bytes after the numbers section, there must
	// only be a single char that determines units
	if i != lenstr-1 {
		return 0, errors.New("invalid format")
	}

	lastChar := str[i]
	switch lastChar {
	case 'B':
		return val, nil
	case 'K':
		return val * 1024, nil
	case 'M':
		return val * 1024 * 1024, nil
	case 'G':
		return val * 1024 * 1024 * 1024, nil
	default:
		return 0, errors.New("invalid units")
	}
}

func (o *Opts) Input() string {
	return *o.input
}

func (o *Opts) WorkingDir() string {
	return *o.wd
}

func (o *Opts) Port() uint16 {
	return uint16(*o.port)
}

func (o *Opts) Cmd() string {
	return *o.cmd
}

func (o *Opts) UpLimit() int64 {
	return o.uplim
}

func (o *Opts) DnLimit() int64 {
	return o.dnlim
}
