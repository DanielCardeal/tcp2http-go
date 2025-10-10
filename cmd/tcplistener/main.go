package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

const SERVER_PORT = ":42069"

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func() {
		defer f.Close()
		defer close(out)

		currentLine := ""
		for {
			buff := make([]byte, 8)
			n, err := f.Read(buff)
			if err != nil {
				break
			}

			buff = buff[:n]
			if i := bytes.IndexByte(buff, '\n'); i != -1 {
				currentLine += string(buff[:i])
				buff = buff[i+1:]
				out <- currentLine
				currentLine = ""

			}
			currentLine += string(buff)
		}
		out <- currentLine
	}()

	return out
}

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

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}
	}
}
