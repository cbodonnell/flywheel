package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/servers"
	"github.com/cbodonnell/flywheel/pkg/state"
	"github.com/cbodonnell/flywheel/pkg/version"
)

func main() {
	// TODO: real logging
	fmt.Printf("Starting server version %s\n", version.Get())
	ctx := context.Background()

	// TODO: don't hard code these
	tcpPort := "8888"
	udpPort := "8889"

	clientManager := clients.NewClientManager()
	messageQueue := queue.NewInMemoryQueue()

	tcpServer := servers.NewTCPServer(clientManager, messageQueue, tcpPort)
	udpServer := servers.NewUDPServer(clientManager, messageQueue, udpPort)
	go tcpServer.Start()
	go udpServer.Start()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable must be set")
	}
	repository := repositories.NewPostgresRepository(ctx, connStr)
	defer repository.Close(ctx)

	stateManager := state.NewInMemoryStateManager()

	gameLoopInterval := 100 * time.Millisecond // 10 FPS
	saveLoopInterval := 5 * time.Second
	gameManager := game.NewGameManager(game.NewGameManagerOptions{
		ClientManager:    clientManager,
		MessageQueue:     messageQueue,
		Repository:       repository,
		StateManager:     stateManager,
		GameLoopInterval: gameLoopInterval,
		SaveLoopInterval: saveLoopInterval,
	})

	fmt.Println("Starting game manager")
	gameManager.Start(ctx)
}
