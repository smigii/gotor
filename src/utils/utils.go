package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const GotorPeerString string = "-GT0000-"

func NewPeerId() string {
	return GotorPeerString + randStringBytesMaskImprSrcSB(12)
}

// Some random bullshit I got from stackoverflow
// I have absolutely no idea what this is doing, but it seems to work
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
	const unit = 1000

	if nbytes < unit {
		return asfloat, "B"
	}

	div, exp := int64(unit), 0
	for n := asfloat / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return asfloat / float64(div), string("KMGTPE"[exp]) + "B"
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
