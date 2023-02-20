package torrentfile

import (
	"Bittorrent-client/bitfield"
	"Bittorrent-client/p2p"
	"Bittorrent-client/peers"
	"bytes"
	"crypto/sha1"
	"fmt"
	"math"
	"os"

	"github.com/jackpal/bencode-go"
)

const Port uint16 = 6087

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type bencodeInfo struct {
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
}

func (tf *TorrentFile) Download(path string) error {
	peerID, err := peers.GeneratePeerID()
	if err != nil {
		return err
	}
	//just hardcode the ip and port number for the seeder: IP:Port and return that as list
	seed, _ := peers.Unmarshal()
	pieceLength := len(tf.PieceHashes)
	ByteSize := 8

	bitfield := make(bitfield.Bitfield, int(math.Ceil(float64(pieceLength)/ float64(ByteSize))))
	outFile, err := os.OpenFile(tf.Name, os.O_RDWR|os.O_CREATE, 0666)

	t := p2p.Torrent{
		Peers:       seed,
		PeerID:      peerID,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
		File: 		 outFile,
		Bitfield:    bitfield,
	}

	er := t.Download()
	if er != nil {
		return er
	}

	return nil
}

func OpenFile(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err

	}
	defer file.Close()

	bct := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bct)
	if err != nil {
		return TorrentFile{}, err
	}
	return bct.toTorrentFile()
}

func WriteFile(path string, buf []byte) error {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(path)
		return err
	}
	defer file.Close()
	_, err = file.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (bct *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, pieceHashes, err := bct.Info.hashInfo()
	if err != nil {
		return TorrentFile{}, err
	}
	return TorrentFile{
		Announce:    bct.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bct.Info.PieceLength,
		Length:      bct.Info.Length,
		Name:        bct.Info.Name,
	}, nil
}

func (bci *bencodeInfo) hashInfo() ([20]byte, [][20]byte, error) {
	pieces := []byte(bci.Pieces)
	hashLen := 20
	numHashes := len(pieces) / hashLen
	pieceHashes := make([][20]byte, numHashes)

	if len(pieces)%hashLen != 0 {
		err := fmt.Errorf("reading hash info failed: invalid hash length (length: %d - expected: %d", len(pieces), hashLen)
		return [20]byte{}, [][20]byte{}, err
	}
	for i := range pieceHashes {
		copy(pieceHashes[i][:], pieces[i*hashLen:(i+1)*hashLen])
	}

	var info bytes.Buffer
	err := bencode.Marshal(&info, *bci)
	if err != nil {
		return [20]byte{}, [][20]byte{}, err
	}
	infoHash := sha1.Sum(info.Bytes())
	return infoHash, pieceHashes, nil
}