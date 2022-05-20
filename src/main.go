package main

import (
	"fmt"
	"gotor/src/bencode"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	execute()
	//quickTest()

	fmt.Println("\n\nDONE")

}

func execute() bool {
	fdata, err := os.ReadFile("src/bencode/test1")
	//fdata, err := os.ReadFile("torrents/ubuntu-20.04.4-desktop-amd64.iso.torrent")
	check(err)

	d, err := bencode.Decode(fdata)
	check(err)

	dict := d.(bencode.Dict)
	fmt.Println(dict["key"])
	return err == nil
}

func quickTest() {

	d, err := bencode.Decode([]byte("i123"))
	if err != nil {
		fmt.Println("ERROR", err)
	} else {
		fmt.Println("SUCCESS", d)
	}

}
