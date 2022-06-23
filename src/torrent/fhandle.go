package torrent

type FileHandler interface {
	// Piece returns the file data at the given piece index.
	Piece(index uint64) ([]byte, error)

	// Write writes the given data to the file(s) corresponding to piece index.
	Write(index uint64, data []byte) error

	// FileMeta returns the metadata for the files in the torrent.
	FileMeta() *TorFileMeta
}
