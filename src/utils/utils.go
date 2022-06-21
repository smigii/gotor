package utils

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const GotorPeerString string = "-GT0000-"

func NewPeerId() string {
	return GotorPeerString + randStringBytesMaskImprSrcSB(12)
}

// Some random bullshit I got from stackoverflow
// I have absolutely no idea how this works, but it seems to work
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func randStringBytesMaskImprSrcSB(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// Bytes4Humans Turns a number of bytes into human-readable number and units.
// Returns the converted number and string representing units
func Bytes4Humans(nbytes uint64) (float64, string) {
	asfloat := float64(nbytes)
	const unit = 1024

	if nbytes < unit {
		return asfloat, "B"
	}

	div, exp := int64(unit), 0
	for n := asfloat / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return asfloat / float64(div), string("KMGTPE"[exp]) + "iB"
}

// Spinner A fun spinner thing, it's very fun
func Spinner(done chan bool, msg string) {
	symbols := []rune{'⠏', '⠛', '⠹', '⢸', '⣰', '⣤', '⣆', '⡇'}
	dots := "..."
	i := 0
	backs := "  \b\b"

	for {
		select {
		case <-done:
			fmt.Printf("\r✓ %v\n", msg)
			return
		default:
			i = (i + 1) % 3
			d := dots[:i+1]
			for _, r := range symbols {
				fmt.Printf("\r%c %v%v%v", r, msg, d, backs)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// SegmentData splits a byte slice into smaller byte slices of size segSize.
// The last segment may have a smaller size than all the other segments.
func SegmentData(data []byte, segSize uint64) [][]byte {
	npieces := (uint64(len(data)) / segSize) + 1
	pieces := make([][]byte, 0, npieces)
	left := uint64(len(data))
	idx := uint64(0)

	for {
		if left == 0 {
			break
		}
		toWrite := segSize
		if left < segSize {
			toWrite = left
		}
		pieces = append(pieces, data[idx:idx+toWrite])
		idx += toWrite
		left -= toWrite
	}
	return pieces
}

// writeSize is the number of bytes written to a file at a time
const writeSize = 1048576

// WriteEmptyFile writes an empty file of specified size.
func WriteEmptyFile(fpath string, size uint64) error {

	e := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
	if e != nil {
		return e
	}

	f, e := os.Create(fpath)
	if e != nil {
		return e
	}

	left := size
	data := make([]byte, writeSize) // 1MiB write
	for {
		if left == 0 {
			break
		}
		if writeSize < left {
			_, e = f.Write(data)
			left -= writeSize
		} else {
			_, e = f.Write(data[:left])
			left = 0
		}
		if e != nil {
			return e
		}
	}

	return nil
}

// CleanUpTestFile will call os.RemoveAll on the base directory specified
// in fpath.
func CleanUpTestFile(fpath string) {
	parts := strings.Split(fpath, "/")
	if len(parts) == 0 {
		fmt.Printf("Could not remove test file [%v]\n", fpath)
		return
	}

	e := os.RemoveAll(parts[0]) // Clean up
	if e == nil {
		fmt.Printf("RemoveAll(%v) successful [input %v]\n", parts[0], fpath)
	} else {
		fmt.Printf("RemoveAll(%v) failed [input %v]\n", parts[0], fpath)
	}
}
