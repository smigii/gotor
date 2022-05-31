package main

import (
	"fmt"
	"gotor/bencode"
	"gotor/torrent"
	"gotor/tracker"
	"gotor/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
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
	req := tracker.NewRequest(tor, 60666)
	fmt.Println("Full URL: ", req.URL)

	go func() {
		defer wg.Done()
		go utils.Spinner(ch, "Downloading peer list from tracker")
	}()
	go func() {
		defer wg.Done()
		//time.Sleep(5 * time.Second)
		resp, err = GetTrackerResponse(req)
		ch <- true
	}()
	wg.Wait()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("\n", resp.Pretty())

	fmt.Println("DONE")

}

func GetTrackerResponse(req *http.Request) (*tracker.Resp, error) {
	client := http.Client{}
	resp, err := client.Do(req)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ben, err := bencode.Decode(body)
	if err != nil {
		return nil, err
	}

	dict, ok := ben.(bencode.Dict)
	if !ok {
		return nil, fmt.Errorf("response not a bencoded dictionary\n%v", body)
	}

	tresp, err := tracker.NewResponse(dict)
	if err != nil {
		return nil, err
	} else {
		return tresp, nil
	}
}
