package fileio

import "gotor/torrent"

type FTree struct {
	root Dir
}

func NewFTree(t *torrent.Torrent) *FTree {

	return nil
}

type Dir struct {
	File
	files []File
	dirs  []Dir
}

type File struct {
	name string
	len  uint64
}
