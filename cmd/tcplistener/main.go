package main

import (
	"fmt"
	"log"
	"net"

	"github.com/DanielCardeal/tcp2http-go/internal/request"
)

const SERVER_PORT = ":42069"

func main() {
	ln, err := net.Listen("tcp", SERVER_PORT)
	if err != nil {
		log.Fatalf("Unable to listen on port %s: %s", SERVER_PORT, err)
	}
	defer ln.Close()

	log.Printf("Server up on port %s", SERVER_PORT)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err)
			continue
		}
		log.Printf("Connection accepted from %s", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("failed to parse http request from %s: %s", conn.RemoteAddr(), err)
			conn.Close()
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		conn.Close()
	}
}
