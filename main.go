package main

import (
	"crypto/rand"
	"fmt"
	"go-torrent/bencode"
	p2p "go-torrent/peers"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	file, err := os.Open("/home/korok/personal/go-torrent/debian-edu-12.2.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	torrent, err := bencode.Open(file)
	if err != nil {
		fmt.Println("Error decoding file:", err)
	}

	torrentFile, err := torrent.ToTorrentFile()
	if err != nil {
		fmt.Println("Error converting torrent file:", err)
	}

	fmt.Println(torrentFile.InfoHash)

	var randomBytes [20]byte
	_, err = rand.Read(randomBytes[:])
	if err != nil {
		// Handle the error appropriately
		log.Fatalf("Failed to generate random bytes: %v", err)
	}

	fmt.Println(torrentFile.BuildTrackerURL(randomBytes, 6881))

	url, err := torrentFile.BuildTrackerURL(randomBytes, 6881)
	if err != nil {
		// Handle the error appropriately
		log.Fatalf("Failed to build URL tracker random bytes: %v", err)
	}
	body, err := get(url)
	if err != nil {
		// Handle the error appropriately
		log.Fatalf("Failed to HTTP get bytes: %v", err)
	}

	res, err := bencode.Unmarshal(body)
	if err != nil {
		panic(err)
	}

	peers, err := p2p.Unmarshal([]byte(res.Peers))
	if err != nil {
		panic(err)
	}
	fmt.Println(peers)

	for _, peer := range peers {
		conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(conn)
	}
}

func get(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return http.NoBody, err
	}

	return resp.Body, nil
}
