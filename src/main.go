package main

import (
	"fmt"
	"gotor/torrent"
	"gotor/tracker"
	"gotor/utils"
	"log"
	"os"
	"sync"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: gotor path/to/file.torrent")
	}

	tor, err := torrent.NewTorrent(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Torrent Info")
	fmt.Println(tor.QuickStats())

	wg := sync.WaitGroup{}
	wg.Add(2)

	var resp *tracker.Resp
	ch := make(chan bool)

	go func() {
		defer wg.Done()
		go utils.Spinner(ch, "Downloading peer list from tracker")
	}()
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Second)
		resp, err = tracker.Request(tor)
		ch <- true
	}()
	wg.Wait()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(resp.Interval())

	fmt.Println("DONE")

}
