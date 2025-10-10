package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const SERVER_ADDR = ":42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", SERVER_ADDR)
	if err != nil {
		log.Fatalf("Unable to resolve UDP address %s", SERVER_ADDR)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Unable to connect to UDP address %s", udpAddr)
	}
	defer conn.Close()

	stdinReader :=  bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			log.Printf("Failed to read stdin: %s", err)
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("Failed to write to UDP connection: %s", err)
		}
	}
}
