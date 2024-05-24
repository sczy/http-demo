package main

import (
	"crypto/tls"
	"log"
	"net"
)

const (
	preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, len(preface))
	_, err := conn.Read(buf)
	if err != nil || string(buf) != preface {
		log.Println("Invalid preface")
		return
	}

	// Basic HTTP/2 settings frame
	settings := []byte{
		0x00, 0x00, 0x12, // Length
		0x04,                   // Type: SETTINGS
		0x00,                   // Flags
		0x00, 0x00, 0x00, 0x00, // Stream ID
		0x00, 0x03, 0x00, 0x00, 0x01, // SETTINGS_MAX_CONCURRENT_STREAMS
		0x00, 0x04, 0x00, 0x00, 0x80, // SETTINGS_INITIAL_WINDOW_SIZE
	}
	conn.Write(settings)

	// Simplified example - reading frames and responding with a single DATA frame
	for {
		frameHeader := make([]byte, 9)
		_, err := conn.Read(frameHeader)
		if err != nil {
			return
		}

		length := int(frameHeader[0])<<16 | int(frameHeader[1])<<8 | int(frameHeader[2])
		frameType := frameHeader[3]
		streamID := int(frameHeader[5])<<24 | int(frameHeader[6])<<16 | int(frameHeader[7])<<8 | int(frameHeader[8])

		payload := make([]byte, length)
		_, err = conn.Read(payload)
		if err != nil {
			return
		}

		if frameType == 0x01 { // HEADERS frame
			// Respond with a DATA frame
			data := []byte("Hello, HTTP/2 with multiplexing!")
			dataFrame := make([]byte, 9+len(data))
			dataFrame[0] = byte(len(data) >> 16)
			dataFrame[1] = byte(len(data) >> 8)
			dataFrame[2] = byte(len(data))
			dataFrame[3] = 0x00 // Type: DATA
			dataFrame[4] = 0x01 // Flags: END_STREAM
			dataFrame[5] = byte(streamID >> 24)
			dataFrame[6] = byte(streamID >> 16)
			dataFrame[7] = byte(streamID >> 8)
			dataFrame[8] = byte(streamID)
			copy(dataFrame[9:], data)
			conn.Write(dataFrame)
		}
	}
}

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2"},
	}

	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer listener.Close()

	log.Println("Starting HTTP/2 server on https://localhost:8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err)
			continue
		}
		go handleConnection(conn)
	}
}
