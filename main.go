package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal("Failed to open messages.txt")
	}

	readBuff := make([]byte, 8)
	currentLine := ""
	n, err := file.Read(readBuff)
	for err != io.EOF && n != 0 {
		segments := strings.Split(string(readBuff[:n]), "\n")
		currentLine += segments[0]
		if len(segments) > 1 {
			fmt.Printf("read: %s\n", currentLine)
			var i int
			for i = 1; i < len(segments)-1; i++ {
				fmt.Printf("read: %s\n", segments[i])
			}
			currentLine = segments[i]
		}
		n, err = file.Read(readBuff)
	}
}
