package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strings"
)

const (
	frameHeaderLen = 9
)

type Frame struct {
	Length   uint32
	Type     uint8
	Flags    uint8
	StreamID uint32
	Payload  []byte
}

func readFrame(conn net.Conn) (*Frame, error) {
	header := make([]byte, frameHeaderLen)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}

	length := uint32(header[0])<<16 | uint32(header[1])<<8 | uint32(header[2])
	frameType := uint8(header[3])
	flags := uint8(header[4])
	streamID := binary.BigEndian.Uint32(header[5:]) & 0x7FFFFFFF

	payload := make([]byte, length)
	_, err = io.ReadFull(conn, payload)
	if err != nil {
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

func writeFrame(conn net.Conn, frame *Frame) error {
	header := make([]byte, frameHeaderLen)
	header[0] = byte(frame.Length >> 16)
	header[1] = byte(frame.Length >> 8)
	header[2] = byte(frame.Length)
	header[3] = frame.Type
	header[4] = frame.Flags
	binary.BigEndian.PutUint32(header[5:], frame.StreamID&0x7FFFFFFF)

	_, err := conn.Write(header)
	if err != nil {
		return err
	}
	_, err = conn.Write(frame.Payload)
	return err
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read initial client preface (24 bytes)
	preface := make([]byte, 24)
	_, err := io.ReadFull(conn, preface)
	if err != nil {
		log.Println("Error reading preface:", err)
		return
	}

	if string(preface) != "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n" {
		log.Println("Invalid preface:", string(preface))
		return
	}

	log.Println("Received valid client preface")

	// Send initial SETTINGS frame
	initialSettings := &Frame{
		Length:   0,
		Type:     4, // SETTINGS
		Flags:    0,
		StreamID: 0,
		Payload:  []byte{},
	}
	err = writeFrame(conn, initialSettings)
	if err != nil {
		log.Println("Error writing initial settings frame:", err)
		return
	}

	log.Println("Sent initial SETTINGS frame")

	for {
		frame, err := readFrame(conn)
		if err != nil {
			log.Println("Error reading frame:", err)
			break
		}

		switch frame.Type {
		case 0x1: // HEADERS
			log.Println("Received HEADERS frame")
			headers := parseHeaders(frame.Payload)
			log.Println("Received headers:", headers)
			// Send a simple response (HEADERS + DATA frames)
			sendSimpleResponse(conn, frame.StreamID)
		case 0x4: // SETTINGS
			log.Println("Received SETTINGS frame")
			if frame.Flags&0x1 == 0 {
				ackSettingsFrame := &Frame{
					Length:   0,
					Type:     4, // SETTINGS
					Flags:    1, // ACK
					StreamID: 0,
					Payload:  []byte{},
				}
				err = writeFrame(conn, ackSettingsFrame)
				if err != nil {
					log.Println("Error writing ACK for SETTINGS frame:", err)
				}
			}
		case 0x8: // WINDOW_UPDATE
			log.Println("Received WINDOW_UPDATE frame")
			// Handle window update if needed
		default:
			log.Printf("Received frame type %d\n", frame.Type)
		}
	}
}

func parseHeaders(payload []byte) map[string]string {
	// Simplified header parsing (actual implementation requires HPACK decoding)
	headers := make(map[string]string)
	parts := strings.Split(string(payload), "\x00")
	for i := 0; i < len(parts)-1; i += 2 {
		headers[parts[i]] = parts[i+1]
	}
	return headers
}

func sendSimpleResponse(conn net.Conn, streamID uint32) {
	// HEADERS frame
	headersPayload := []byte{
		0x88,       // :status: 200 (pre-encoded in HPACK format)
		0x40, 0x0c, // content-type
		0x74, 0x65, 0x78, 0x74, // "text"
		0x2f, 0x70, 0x6c, 0x61, // "/pla"
		0x69, 0x6e, 0x3b, 0x20, // "in; "
		0x63, 0x68, 0x61, 0x72, // "char"
		0x73, 0x65, 0x74, 0x3d, // "set="
		0x75, 0x74, 0x66, 0x2d, // "utf-"
		0x38, // "8"
	}
	headersFrame := &Frame{
		Length:   uint32(len(headersPayload)),
		Type:     1, // HEADERS
		Flags:    4, // END_HEADERS
		StreamID: streamID,
		Payload:  headersPayload,
	}
	err := writeFrame(conn, headersFrame)
	if err != nil {
		log.Println("Error writing headers frame:", err)
		return
	}

	// DATA frame
	data := []byte("Hello, HTTP/2!")
	dataFrame := &Frame{
		Length:   uint32(len(data)),
		Type:     0, // DATA
		Flags:    1, // END_STREAM
		StreamID: streamID,
		Payload:  data,
	}
	err = writeFrame(conn, dataFrame)
	if err != nil {
		log.Println("Error writing data frame:", err)
	}
}

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := tls.Listen("tcp", ":8080", config)
	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}
	defer listener.Close()

	log.Println("Listening on https://localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
