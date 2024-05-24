package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

const (
	preface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

func sendRequest(conn net.Conn, streamID int) {
	// Simplified HTTP/2 HEADERS frame
	headers := []byte{
		0x00, 0x00, 0x00, // Length
		0x01, // Type: HEADERS
		0x05, // Flags: END_HEADERS | END_STREAM
		byte(streamID >> 24), byte(streamID >> 16), byte(streamID >> 8), byte(streamID),
	}
	conn.Write(headers)

	// Read response DATA frame
	frameHeader := make([]byte, 9)
	_, err := conn.Read(frameHeader)
	if err != nil {
		log.Printf("failed to read frame header: %s", err)
		return
	}

	length := int(frameHeader[0])<<16 | int(frameHeader[1])<<8 | int(frameHeader[2])
	payload := make([]byte, length)
	_, err = conn.Read(payload)
	if err != nil {
		log.Printf("failed to read frame payload: %s", err)
		return
	}

	fmt.Printf("Response for stream %d: %s\n", streamID, payload)
}

func main() {
	// Load server certificate to trust
	certPool := x509.NewCertPool()
	certBytes, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatalf("failed to read server certificate: %s", err)
	}
	certPool.AppendCertsFromPEM(certBytes)

	config := &tls.Config{
		RootCAs:    certPool,
		NextProtos: []string{"h2"},
	}

	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer conn.Close()

	// Send client preface
	conn.Write([]byte(preface))

	// Read server settings frame
	frameHeader := make([]byte, 9)
	_, err = conn.Read(frameHeader)
	if err != nil {
		log.Fatalf("failed to read frame header: %s", err)
	}

	length := int(frameHeader[0])<<16 | int(frameHeader[1])<<8 | int(frameHeader[2])
	payload := make([]byte, length)
	_, err = conn.Read(payload)
	if err != nil {
		log.Fatalf("failed to read frame payload: %s", err)
	}

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(streamID int) {
			defer wg.Done()
			sendRequest(conn, streamID)
		}(i)
	}

	wg.Wait()
}
