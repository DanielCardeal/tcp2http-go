package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal("Failed to open messages.txt")
	}

	msgBuffer := make([]byte, 8)
	n, err := file.Read(msgBuffer)
	for err != io.EOF {
		fmt.Printf("read: %s\n", msgBuffer[:n])
		n, err = file.Read(msgBuffer)
	}
}
