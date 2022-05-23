package main

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
