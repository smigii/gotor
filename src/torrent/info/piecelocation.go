package info

import "gotor/torrent/filesd"

type PieceLocation struct {
	Entry *filesd.Entry
	Loc   filesd.PieceInfo
}

func (pl PieceLocation) Path() string {
	return pl.Entry.LocalPath()
}

func (pl PieceLocation) ReadAmnt() int64 {
	return pl.Loc.ReadAmnt
}

func (pl PieceLocation) SeekAmnt() int64 {
	return pl.Loc.SeekAmnt
}
