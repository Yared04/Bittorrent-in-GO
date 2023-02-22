package main

import (
	"log"
	"os"
	"Bittorrent-client/torrentfile"
)

func main() {
	input_torrent := os.Args[1]
	output_file := os.Args[2]

	torrent_file, err := torrentfile.OpenFile(input_torrent)
	if err != nil {
		log.Fatal(err)
	}

	err = torrent_file.Download(output_file)
	if err != nil {
		log.Fatal(err)
	}
}
