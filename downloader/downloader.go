package downloader

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"Bittorrent-client/seeder"
	"Bittorrent-client/message"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

// Torrent holds data required to download a torrent from a list of peers
type Torrent struct {
	Peers      []seeder.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	File 		*os.File
	Bitfield 	message.Bitfield
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *seeder.Seeder
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read() // this call blocks
	if err != nil {
		return err
	}

	if msg == nil { // keep-alive
		return nil
	}

	if msg.ID == message.MsgPiece {
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownloadPiece(c *seeder.Seeder, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	// Setting a deadline helps get unresponsive peers unstuck.
	c.Conn.SetDeadline(time.Now().Add(20 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pw.length {
		// If unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloader(peer seeder.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := seeder.Connect(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw // Put piece back on the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw // Put piece back on the queue
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw // Put piece back on the queue
			continue
		}

		results <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

// Download downloads the torrent. This stores the downloaded file in the current directory.
func (t *Torrent) Download() error {
	log.Println("Starting download for", t.Name)
	// Init queues for workers to retrieve work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	completePieces := 0
	restore := false
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		start, _ := t.calculateBoundsForPiece(index)
		pieceWork := pieceWork{index, hash, length}

		SinglePiece := make([]byte , length)
		t.File.ReadAt(SinglePiece, int64(start))

		integrity_error := checkIntegrity(&pieceWork, SinglePiece)

		if integrity_error != nil {
			workQueue <- &pieceWork
		} else {
			restore = true
			completePieces ++
			t.Bitfield.SetPiece(index)
		}


		
	}

	// Start workers
	for _, peer := range t.Peers {
		go t.startDownloader(peer, workQueue, results)
	}

	if (restore) {
		log.Printf("Download is Resuming from:  %0.2f%%", float64(completePieces) / float64(len(t.PieceHashes)) * 100)
		
	}
	for completePieces < len(t.PieceHashes) {
		res := <-results
		begin, _ := t.calculateBoundsForPiece(res.index)
		
		//write to result
		_, err := t.File.WriteAt(res.buf, int64(begin))
		if err != nil{
			log.Fatal("Failed to write file to memory")
		}
		completePieces++

		percent := float64(completePieces) / float64(len(t.PieceHashes)) * 100
		// numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread

		fmt.Printf("\r Download Progress: %s  %0.2f%%",strings.Repeat("#", int(percent/10)), percent)

	}

	log.Printf("Download Finished!!!, You can find your downloaded file in the current directory.")
	close(workQueue)

	return nil
}
