package main

import (
	"Bittorrent-client/handshake"
	"Bittorrent-client/message"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

const (
	HOST = "127.0.0.1"
	PORT = "8080"
	TYPE = "tcp"


)

type Torrent struct {
	Bitfield []byte `json:"bitfield"`
	Path string `json:"path"`
	Piecelength float64 `json:"piecelength"`
	Length float64 `json:"length"`
}

func main() {
	listen, err := net.Listen(TYPE, ":"+PORT)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(fmt.Sprintf("listening on %s:%s", HOST, PORT))
	// close listener
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}




func handleRequest(conn net.Conn) {
	
	defer conn.Close()
		
	torrent, err := handshakeSeeder(conn)
	if err != nil{
		
		return 
	}
	

	sendBitfeild(torrent, conn)

	piecelength := torrent.Piecelength
  	
	f, err := os.Open(torrent.Path)

	if err != nil {
        fmt.Println("File reading error", err)
        return
    }

	for {
		msg, err := message.Read(conn)
		
		if err != nil {
			return
		}

		go serveRequest(msg, piecelength, f, conn )
	}

	
}
func serveRequest(msg *message.Message, piecelength float64, file *os.File, connection net.Conn) {
	if msg.ID == message.MsgRequest {
			
		index, begin, length := binary.BigEndian.Uint32(msg.Payload[0:4]), binary.BigEndian.Uint32(msg.Payload[4:8]), binary.BigEndian.Uint32(msg.Payload[8:])
		
		content := make([]byte, length) 
		
		_, err := file.ReadAt(content, int64(int64(index)*int64(piecelength) + int64(begin)))
		if err != nil{
			log.Fatal("Error Reading File.")
		}
		piece := getPiece(content, index, begin, length)
		connection.Write(piece.Serialize())
	}
}

func getPiece(content []byte, index uint32, begin uint32, length uint32) *message.Message{

	buf := make([]byte, 8 + length)
	binary.BigEndian.PutUint32(buf[0:4], uint32(index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(begin))
	copy(buf[8:], content)
	msg := &message.Message{
		ID:      message.MsgPiece,
		Payload: buf,
	}
	
	return msg
	
}


func handshakeSeeder(conn net.Conn) (*Torrent, error){
	res, err := handshake.Read(conn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}


	infoHash := hex.EncodeToString(res.InfoHash[:])

	// Open our torrent jsonFile
	jsonFile, err := os.Open(fmt.Sprintf("files/%s.json", infoHash))
	defer jsonFile.Close()
	if err != nil {
		log.Printf("Recived Request for file with hash %s but file is not available", infoHash )
		return nil, err
	}

	torrent := Torrent{}

	byteValue, err := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &torrent)
	
	conn.Write(res.Serialize())

	return &torrent, err
}

func sendBitfeild(torrent *Torrent, conn net.Conn){

	var bf message.Bitfield = []byte(torrent.Bitfield)
	msg := &message.Message{
		ID:      message.MsgBitfield,
		Payload: bf,
	}
	conn.Write(msg.Serialize())

}

