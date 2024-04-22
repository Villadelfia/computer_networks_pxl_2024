package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
)

func main() {
	// Parse flags
	addr := flag.String("addr", "0.0.0.0", "The address to listen on")
	port := flag.Int("port", 5555, "The port to listen on")
	flag.Parse()

	// Build the connection string
	conn := fmt.Sprintf("%s:%d", *addr, *port)

	// Set up listener
	listener, err := net.Listen("tcp", conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listening on %s...\nPress CTRL+C to quit.\nStart the server with --addr=<address> and --port=<port> to change the default address and port.\n", conn)

	// And just infinite loop waiting for connections
	count := 0
	for {
		// We got a new connection, try and establish it
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Connection established, handle it
		go handleClient(conn, count)

		// Increment the connection count, for debug output
		count++
	}
}

func handleClient(conn net.Conn, id int) {
	defer conn.Close()

	// Generate a random number from 0 to 1000000
	r := rand.Intn(1000001)
	fmt.Printf("[Client %d] Random number to guess: %d\n", id, r)

	// Get a buffered reader and writer
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Loop until guess is correct
	guess := -1
	for guess != r {
		// Read the guess
		received := make([]byte, 4)
		_, err := io.ReadFull(reader, received)
		if err != nil {
			fmt.Printf("[Client %d] Error: %s\n", id, err)
			return
		}
		guess := int(binary.BigEndian.Uint32(received))

		// Print the guess
		fmt.Printf("[Client %d] Received guess: %d\n", id, guess)

		// Send the response
		if guess < r {
			fmt.Printf("[Client %d] Sending \"Higher\"\n", id)
			_, err := writer.WriteString("Higher")
			if err != nil {
				fmt.Printf("[Client %d] Error: %s\n", id, err)
				return
			}
			writer.Flush()
		} else if guess > r {
			fmt.Printf("[Client %d] Sending \"Lower\"\n", id)
			_, err := writer.WriteString("Lower")
			if err != nil {
				fmt.Printf("[Client %d] Error: %s\n", id, err)
				return
			}
			writer.Flush()
		} else {
			fmt.Printf("[Client %d] Sending \"Correct\"\n", id)
			_, err := writer.WriteString("Correct")
			if err != nil {
				fmt.Printf("[Client %d] Error: %s\n", id, err)
				return
			} else {
				writer.Flush()
				break
			}
		}
	}

	// We're done
	fmt.Printf("[Client %d] Closing connection\n", id)
}
