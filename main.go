package main

import (
	"fmt"
	"gotor/bencode"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	//fdata, err := os.ReadFile("bencode/test1")
	fdata, err := os.ReadFile("torrents/ubuntu-20.04.4-desktop-amd64.iso.torrent")
	check(err)

	d, err := bencode.Decode(fdata)
	check(err)

	//dict := map[string]interface{}(d)

	fmt.Println(d)

}

func execute() bool {
	fdata, err := os.ReadFile("torrents/ubuntu-20.04.4-desktop-amd64.iso.torrent")
	check(err)

	_, err = bencode.Decode(fdata)
	return err == nil
}
