package fileio

import (
	"gotor/bf"
	"gotor/utils"
)

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
	// Piece reads the file data for the given index into the buf byte slice.
	// Returns the number of bytes read.
	Piece(index int64, buf []byte) (int64, error)

	// Write writes the given data to the file(s) corresponding to piece index.
	// It will only write data if the SHA1 hash of the data matches the hash
	// given in the meta's hash string. If it does not match, a HashError will
	// be returned.
	Write(index int64, data []byte) error

	// FileMeta returns the metadata for the files in the torrent.
	FileMeta() *TorFileMeta

	// Validate will look through the file(s) specified in the torrent and
	// check the pieces and their hashes, updating the results in its bitfield.
	Validate() error

	Bitfield() *bf.Bitfield

	Close() error
}

// ValidateHandler will probably be deleted. Will write a more optimized
// FileHandler.Validate() for SingleFileHandler and MultiFileHandler
// that reads in larger chunks of data at a time.
func ValidateHandler(fh FileHandler) error {
	buf := make([]byte, fh.FileMeta().PieceLen(), fh.FileMeta().PieceLen())

	var i int64
	for i = 0; i < fh.FileMeta().numPieces; i++ {

		knownHash, e := fh.FileMeta().PieceHash(i)
		if e != nil {
			return e
		}

		n, e := fh.Piece(i, buf)
		if e != nil {
			return e
		}

		val := utils.SHA1(buf[:n]) == knownHash
		fh.Bitfield().Set(i, val)
	}

	return nil
}
