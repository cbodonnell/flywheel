package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/servers"
)

func main() {
	tcpPort := "8888"
	udpPort := "8889"

	clientManager := clients.NewClientManager()
	messageQueue := queue.NewMemoryQueue()

	tcpServer := servers.NewTCPServer(clientManager, messageQueue, tcpPort)
	udpServer := servers.NewUDPServer(clientManager, messageQueue, udpPort)
	go tcpServer.Start()
	go udpServer.Start()

	// Start the game loop
	gameManager := game.NewGameManager(clientManager, messageQueue, 100*time.Millisecond)
	go gameManager.StartGameLoop()

	fmt.Println("Server started.")

	// Gracefully handle Ctrl+C to stop the program
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	<-stopSignal

	// Perform cleanup or other graceful shutdown tasks here

	fmt.Println("Shutting down...")
}
