package torrent

import "gotor/utils"

// ============================================================================
// STRUCT =====================================================================

// FileSingle is used for single-file torrents. It can access pieces faster
// than FileList since it doesn't need to find out which files are contained
// int a given piece.
type FileSingle struct {
	meta *TorFileMeta
}

// ============================================================================
// GETTER =====================================================================

func (f *FileSingle) Piece(index uint64) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FileSingle) Write(index uint64, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (f *FileSingle) FileMeta() *TorFileMeta {
	return f.meta
}

func (f *FileSingle) Validate(bf *utils.Bitfield) {
	//TODO implement me
}

// ============================================================================
// CONSTRUCTOR ================================================================

func newFileSingle(meta *TorFileMeta) *FileSingle {
	return &FileSingle{
		meta: meta,
	}
}
