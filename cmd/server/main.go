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
	"github.com/cbodonnell/flywheel/pkg/version"
)

func main() {
	// TODO: real logging
	fmt.Printf("Starting server version %s\n", version.Get())

	tcpPort := "8888"
	udpPort := "8889"

	clientManager := clients.NewClientManager()
	messageQueue := queue.NewMemoryQueue()

	tcpServer := servers.NewTCPServer(clientManager, messageQueue, tcpPort)
	udpServer := servers.NewUDPServer(clientManager, messageQueue, udpPort)
	go tcpServer.Start()
	go udpServer.Start()

	gameLoopInterval := 100 * time.Millisecond // 10 FPS
	gameManager := game.NewGameManager(clientManager, messageQueue, gameLoopInterval)
	go gameManager.StartGameLoop()

	// Gracefully handle Ctrl+C to stop the program
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	<-stopSignal

	// Perform cleanup or other graceful shutdown tasks here

	fmt.Println("Shutting down...")
}
