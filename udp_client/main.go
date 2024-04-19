package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

func main() {
	// Parse flags
	addr := flag.String("addr", "127.0.0.1", "The address to connect to")
	port := flag.Int("port", 5555, "The port to listen on")
	flag.Parse()

	// Build the connection string
	conn := fmt.Sprintf("%s:%d", *addr, *port)

	// Get an address from that
	udpaddr, err := net.ResolveUDPAddr("udp", conn)

	// Set up connection
	connection, err := net.DialUDP("udp", nil, udpaddr)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer connection.Close()
	fmt.Printf("Connected to %s...\nStart the client with --addr=<address> and --port=<port> to change the default address and port.\nYou may keep making guesses!\n> ", conn)

	// Just print all responses as they come
	go getResponses(connection)

	// Timeout stuff
	currenttzero := time.Now().UnixNano()
	currentdeltad, err := time.ParseDuration("16s")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	currentdelta := currentdeltad.Nanoseconds()

	// Loop until the connection is closed
	guess := 0
	for {
		// Check for timeout
		if time.Now().UnixNano()-currenttzero > currentdelta {
			fmt.Print("You lost ?\n> ")
		}

		// Get a guess
		_, err = fmt.Scanln(&guess)
		if err != nil {
			fmt.Printf("Error: %s\n> ", err)
			continue
		}

		// Turn the guess into a string
		strguess := fmt.Sprintf("%d", guess)

		// Send the guess
		n, err := connection.Write([]byte(strguess))
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		fmt.Println(n)

		// Timeout
		currenttzero = time.Now().UnixNano()
		fmt.Print("> ")
	}
}

func getResponses(connection *net.UDPConn) {
	for {
		// Get the response
		response := make([]byte, 256)
		recvd, err := connection.Read(response)
		if err != nil {
			continue
		}
		fmt.Print(string(response[:recvd]), "\n> ")
	}
}
