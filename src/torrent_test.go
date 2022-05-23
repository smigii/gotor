package main

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestTorrent(t *testing.T) {

	err := testTorrent(
		"../torrents/flowers.torrent",
		"https://sometrackerthatdoesntexist.com/announce",
		"9ac55b9c736b1f97d510d7c53c7b6210421cbd06",
	)

	if err != nil {
		t.Error(err)
	}

	err = testTorrent(
		"../torrents/ubuntu-20.04.4-desktop-amd64.iso.torrent",
		"https://torrent.ubuntu.com/announce",
		"f09c8d0884590088f4004e010a928f8b6178c2fd",
	)

	if err != nil {
		t.Error(err)
	}
}

func testTorrent(path string, announce string, infohash string) error {
	tor, err := NewTorrent(path)

	if err != nil {
		return err
	}

	if tor.Announce != announce {
		return fmt.Errorf("bad Announce\nexpected [%v]\ngot      [%v]", announce, tor.Announce)
	}

	trueHashBytes, _ := hex.DecodeString(infohash)
	trueHashString := string(trueHashBytes)

	if tor.Infohash != trueHashString {
		return fmt.Errorf("bad Infohash\nexpected [%v]\ngot      [%v]", infohash, hex.EncodeToString([]byte(tor.Infohash)))
	}

	return nil
}
