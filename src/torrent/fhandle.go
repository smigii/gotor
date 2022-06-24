package torrent

import "gotor/utils"

/* ============================================================================
TODO: Optimize file IO operations through batching
Torrents can easily hit 1k+ pieces, which translates to 1k+ file IO ops.
This will slow us down a lot, we should implement a batching system to
reduce the number of IO requests.
============================================================================ */

type HashError struct{}

func (he *HashError) Error() string {
	return "bad data hash"
}

type FileHandler interface {
	// Piece returns the file data at the given piece index.
	Piece(index int64) ([]byte, error)

	// Write writes the given data to the file(s) corresponding to piece index.
	Write(index int64, data []byte) error

	// FileMeta returns the metadata for the files in the torrent.
	FileMeta() *TorFileMeta

	// Validate will look through the file(s) specified in the torrent and check
	// the pieces and their hashes. If a file doesn't exist, the file will be
	// created and set to the correct size. If a file exists, but is the wrong
	// size, empty bytes will be appended to the correct size. It will overwrite
	// the provided bitfield.
	Validate(bf *utils.Bitfield)
}
