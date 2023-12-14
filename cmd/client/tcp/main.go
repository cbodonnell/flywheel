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
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("Error connecting to TCP server:", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(conn net.Conn, cancel context.CancelFunc) {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			response := scanner.Text()
			fmt.Println("Server:", response)
		}

		fmt.Println("TCP server disconnected.")
		cancel()
	}(conn, cancel)

	go func(conn net.Conn, cancel context.CancelFunc) {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("Enter message (type 'exit' to quit): ")
			scanner.Scan()
			message := scanner.Text()

			_, err := fmt.Fprintf(conn, message+"\n")
			if err != nil {
				fmt.Println("Error sending message to TCP server:", err)
				return
			}

			if message == "exit" {
				fmt.Println("Received exit command, exiting.")
				cancel()
				break
			}
		}
	}(conn, cancel)

	// Gracefully handle Ctrl+C to stop the program
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stopSignal:
		fmt.Println("Received stop signal, exiting.")
	case <-ctx.Done():
	}

	fmt.Println("Exiting TCP client.")
}
