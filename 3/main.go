package main

import (
	"io"
	"log"
	"net"
	"sync"
)

// curl --http2-prior-knowledge -v http://localhost:8080

const (
	Preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

type Frame struct {
	Length   int
	Type     byte
	Flags    byte
	StreamID int
	Payload  []byte
}

func readFrame(conn net.Conn) (*Frame, error) {
	header := make([]byte, 9)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	length := int(header[0])<<16 | int(header[1])<<8 | int(header[2])
	frameType := header[3]
	flags := header[4]
	streamID := int(header[5])<<24 | int(header[6])<<16 | int(header[7])<<8 | int(header[8])

	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	return &Frame{
		Length:   length,
		Type:     frameType,
		Flags:    flags,
		StreamID: streamID,
		Payload:  payload,
	}, nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, len(Preface))
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Printf("Error reading preface: %v", err)
		return
	}
	if string(buf) != Preface {
		log.Printf("Invalid preface: %s", string(buf))
		return
	}

	log.Println("Received valid HTTP/2 preface")

	// Send a simple SETTINGS frame
	settings := []byte{
		0x00, 0x00, 0x00, // Length
		0x04,                   // Type (SETTINGS)
		0x00,                   // Flags
		0x00, 0x00, 0x00, 0x00, // Stream ID
	}
	conn.Write(settings)

	var wg sync.WaitGroup
	for {
		frame, err := readFrame(conn)
		if err != nil {
			log.Printf("Error reading frame: %v", err)
			return
		}

		wg.Add(1)
		go func(f *Frame) {
			defer wg.Done()
			handleFrame(conn, f)
		}(frame)
	}

	wg.Wait()
}

func handleFrame(conn net.Conn, frame *Frame) {
	log.Printf("Received frame: %+v", frame)

	if frame.Type == 0x01 { // HEADERS frame
		response := []byte{
			0x00, 0x00, 0x0C, // Length
			0x01, // Type (HEADERS)
			0x04, // Flags (END_HEADERS)
			byte(frame.StreamID >> 24 & 0xFF), byte(frame.StreamID >> 16 & 0xFF), byte(frame.StreamID >> 8 & 0xFF), byte(frame.StreamID & 0xFF),
			0x88,       // :status 200
			0x0A,       // Content-Length: 10
			0x00, 0x0A, // Literal Header Field without Indexing - Indexed Name
			0x00, 0x00, 0x00, 0x00, 0x00, 0x0A,
		}
		conn.Write(response)

		data := []byte{
			0x00, 0x00, 0x0A, // Length
			0x00, // Type (DATA)
			0x01, // Flags (END_STREAM)
			byte(frame.StreamID >> 24 & 0xFF), byte(frame.StreamID >> 16 & 0xFF), byte(frame.StreamID >> 8 & 0xFF), byte(frame.StreamID & 0xFF),
			'H', 'e', 'l', 'l', 'o', ',', ' ', 'H', '2', '!',
		}
		conn.Write(data)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Println("Listening on :8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
