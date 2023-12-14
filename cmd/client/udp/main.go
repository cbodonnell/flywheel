package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8889")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error connecting to UDP server:", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(conn *net.UDPConn, ctx context.Context) {
		buffer := make([]byte, 1024)

		for {
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				// If the context is Done, then an error is expected
				if ctx.Err() != nil {
					return
				}
				fmt.Println("Error reading UDP message:", err)
				return
			}

			message := string(buffer[:n])
			fmt.Println("Received UDP message:", message)
		}
	}(conn, ctx)

	go func(conn *net.UDPConn, cancel context.CancelFunc) {
		scanner := bufio.NewScanner(os.Stdin)

		for {
			fmt.Print("Enter message (type 'exit' to quit): ")
			scanner.Scan()
			message := scanner.Text()

			if message == "exit" {
				fmt.Println("Received exit command, exiting.")
				cancel()
				break
			}

			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println("Error sending message to UDP server:", err)
				return
			}
		}
	}(conn, cancel)

	// Gracefully handle Ctrl+C to stop the program
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stopSignal:
		fmt.Println("Received stop signal, exiting.")
		cancel()
	case <-ctx.Done():
	}

	fmt.Println("Exiting UDP client.")
}
