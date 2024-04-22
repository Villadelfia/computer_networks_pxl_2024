package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"strings"
)

func main() {
	// Parse flags
	addr := flag.String("addr", "127.0.0.1", "The address to connect to")
	port := flag.Int("port", 5555, "The port to listen on")
	flag.Parse()

	// Build the connection string
	conn := fmt.Sprintf("%s:%d", *addr, *port)

	// Set up listener
	connection, err := net.Dial("tcp", conn)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer connection.Close()
	fmt.Printf("Connected to %s...\nStart the client with --addr=<address> and --port=<port> to change the default address and port.\nMake a guess!\n> ", conn)

	// Loop until the connection is closed
	guess := 0
	for {
		// Get a guess
		_, err := fmt.Scanln(&guess)
		if err != nil {
			fmt.Printf("Error: %s\n> ", err)
			continue
		}

		// Send the guess
		buffer := make([]byte, 4)
		binary.BigEndian.PutUint32(buffer, uint32(guess))
		sent, err := connection.Write(buffer)
		if err != nil || sent != 4 {
			fmt.Println("Error: ", err)
			return
		}

		// Get the response
		response := make([]byte, 256)
		recvd, err := connection.Read(response)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}

		if string(response[:recvd]) == "Correct" {
			fmt.Println("You guessed correctly!")
			return
		} else {
			fmt.Printf("Try guessing %s.\n> ", strings.ToLower(string(response)))
		}
	}
}
