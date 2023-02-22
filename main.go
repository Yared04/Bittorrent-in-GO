package main

// import (
//     "fmt"
//     "net"
// )

// func main() {
//     conn, err := net.Dial("tcp", "192.168.43.30:8080")
//     if err != nil {
//         fmt.Println("Error connecting:", err.Error())
//         return
//     }
//     defer conn.Close()

//     message := "Hello, server!"
//     _, err = conn.Write([]byte(message))
//     if err != nil {
//         fmt.Println("Error sending:", err.Error())
//         return
//     }

//     response := make([]byte, 1024)
//     n, err := conn.Read(response)
//     if err != nil {
//         fmt.Println("Error receiving:", err.Error())
//         return
//     }

//     fmt.Println("Response:", string(response[:n]))
// }


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
