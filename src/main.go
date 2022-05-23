package main

import (
	"fmt"
	"gotor/bencode"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: gotor path/to/file.torrent")
	}

	tor, err := NewTorrent(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(tor.QuickStats())

	// https://torrent.ubuntu.com/announce?info_hash=%F0%9C%8D%08%84Y%00%88%F4%00N%01%0A%92%8F%8Bax%C2%FD&peer_id=shf74nfdhas93hlsaf83&port=0&uploaded=0&downloaded=0&left=0
	client := http.Client{}
	req, _ := http.NewRequest("GET", tor.Announce, nil)
	query := req.URL.Query()
	query.Add("info_hash", tor.Infohash)
	query.Add("peer_id", url.QueryEscape("shf74nfdhas93hlsaf83"))
	query.Add("port", "30666")
	query.Add("uploaded", "0")
	query.Add("downloaded", "0")
	query.Add("left", fmt.Sprintf("%v", tor.Length))
	req.URL.RawQuery = query.Encode()
	log.Println(req.URL)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	dict, err := bencode.Decode(body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(dict)

	log.Println("DONE")

}
