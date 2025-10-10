package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const SERVER_PORT = ":42069"

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)

	go func() {
		defer close(c)
		defer f.Close()

		readBuff := make([]byte, 8)
		currentLine := ""
		n, err := f.Read(readBuff)
		for err != io.EOF {
			segments := strings.Split(string(readBuff[:n]), "\n")
			currentLine += segments[0]
			if len(segments) > 1 {
				c <- currentLine
				var i int
				for i = 1; i < len(segments)-1; i++ {
					c <- segments[i]
				}
				currentLine = segments[i]
			}
			n, err = f.Read(readBuff)
		}
	}()

	return c
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
		c := getLinesChannel(conn)
		for line := range c {
			fmt.Println(line)
		}
	}
}
