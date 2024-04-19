package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func main() {
	// Parse flags
	addr := flag.String("addr", "0.0.0.0", "The address to connect to")
	port := flag.Int("port", 5555, "The port to listen on")
	flag.Parse()

	// Build the connection string
	conn := fmt.Sprintf("%s:%d", *addr, *port)

	// Get an address from that
	udpaddr, err := net.ResolveUDPAddr("udp", conn)

	// Set up listener
	listener, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listening on %s...\nPress CTRL+C to quit.\nStart the server with --addr=<address> and --port=<port> to change the default address and port.\n", conn)

	// Small state machine
	const (
		idle = iota
		ingame
		overtime
	)
	currentstate := idle
	currenttzero := time.Now().UnixNano()
	currenttdelta := int64(0)

	// Miscellaneous data
	buffer := make([]byte, 1024)
	target := 0
	closestguess := 0
	var closestguesser *net.UDPAddr = nil

	// Set up the default timeouts
	defaultingametimeoutd, err := time.ParseDuration("8s")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defaultingametimeout := defaultingametimeoutd.Nanoseconds()
	defaultovertimetimeoutd, err := time.ParseDuration("16s")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defaultovertimetimeout := defaultovertimetimeoutd.Nanoseconds()

	// And just infinite loop waiting for packets
	for {
		if currentstate == idle {
			// Try and get a packet...
			listener.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			recvd, addr, err := listener.ReadFromUDP(buffer)
			if err != nil {
				continue
			}

			// We got data
			fmt.Printf("Got a packet from %s\n", addr.String())

			// New target
			target = rand.Intn(101)
			fmt.Printf("New target: %d\n", target)

			// Mark as best guess
			guess, err := strconv.Atoi(string(buffer[:recvd]))
			if err != nil {
				fmt.Println("Error: ", err)
				continue
			}
			closestguess = guess
			closestguesser = addr
			fmt.Printf("Got guess %d from %s\n", guess, addr.String())

			// Tell them that they're currently the best
			listener.WriteToUDP([]byte("You won ?"), addr)

			// Set time out and change state
			currenttzero = time.Now().UnixNano()
			currenttdelta = defaultingametimeout
			currentstate = ingame
			fmt.Printf("State: ingame\n")
		} else if currentstate == ingame {
			now := time.Now().UnixNano()
			if now-currenttzero > currenttdelta {
				// Time out
				fmt.Printf("Time out, notifying winner and going to overtime\n")
				listener.WriteToUDP([]byte("You won !"), closestguesser)
				currentstate = overtime
				currenttzero = time.Now().UnixNano()
				currenttdelta = defaultovertimetimeout
				fmt.Printf("State: overtime\n")
			} else {
				// Try and get a packet...
				listener.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
				recvd, addr, err := listener.ReadFromUDP(buffer)
				if err != nil {
					continue
				}

				// We got data
				fmt.Printf("Got a packet from %s\n", addr.String())

				// Get guess
				guess, err := strconv.Atoi(string(buffer[:recvd]))
				if err != nil {
					fmt.Println("Error: ", err)
					continue
				}
				fmt.Printf("Got guess %d from %s\n", guess, addr.String())

				// Is it the closest?
				difference := guess - target
				if difference < 0 {
					difference = -difference
				}
				closestdifference := closestguess - target
				if closestdifference < 0 {
					closestdifference = -closestdifference
				}
				if difference < closestdifference {
					// New best guess
					closestguess = guess
					closestguesser = addr
					fmt.Printf("New best guess: %d from %s\n", guess, addr.String())
					listener.WriteToUDP([]byte("You won ?"), addr)
				}

				// Reset zero and halve remaining time
				currenttzero = time.Now().UnixNano()
				currenttdelta = currenttdelta / 2
			}
		} else if currentstate == overtime {
			now := time.Now().UnixNano()
			if now-currenttzero > currenttdelta {
				// Time out
				fmt.Printf("Time out, returning to idle\n")
				currentstate = idle
				fmt.Printf("State: idle\n")
			} else {
				// Try and get a packet...
				listener.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
				_, addr, err := listener.ReadFromUDP(buffer)
				if err != nil {
					continue
				}

				// We got data, but it's too late
				fmt.Printf("Got a packet from %s, but it's too late\n", addr.String())
				listener.WriteToUDP([]byte("You lost !"), addr)
			}
		}
	}
}
