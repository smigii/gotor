package utils

import (
	"fmt"
	"time"
)

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
