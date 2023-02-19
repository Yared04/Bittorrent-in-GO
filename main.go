package main

import (
	"log"
	"os"

	"Bittorrent-client/torrentfile"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrentfile.OpenFile(inPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.Download(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
