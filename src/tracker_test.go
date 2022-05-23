package main

import (
	"fmt"
	"testing"
)

func TestTorrent(t *testing.T) {

	// INFO HASH: f09c8d0884590088f4004e010a928f8b6178c2fd
	tor, err := NewTorrent("../torrents/ubuntu-20.04.4-desktop-amd64.iso.torrent")
	if err != nil {
		t.Error(err)
	}

	//fmt.Println("f09c8d0884590088f4004e010a928f8b6178c2fd")
	//fmt.Println(tor)
	fmt.Println(tor.announce)

	x := "hello"
	y := x
	fmt.Printf("%p\n%p\n", &x, &y)

}
