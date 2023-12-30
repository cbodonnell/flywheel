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
	"github.com/cbodonnell/flywheel/pkg/workers"
)

func main() {
	// TODO: real logging
	fmt.Printf("Starting server version %s\n", version.Get())
	ctx := context.Background()

	// TODO: don't hard code these
	tcpPort := "8888"
	udpPort := "8889"

	clientManager := clients.NewClientManager()
	clientMessageQueue := queue.NewInMemoryQueue()

	tcpServer := servers.NewTCPServer(clientManager, clientMessageQueue, tcpPort)
	udpServer := servers.NewUDPServer(clientManager, clientMessageQueue, udpPort)
	go tcpServer.Start()
	go udpServer.Start()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable must be set")
	}
	repository := repositories.NewPostgresRepository(ctx, connStr)
	defer repository.Close(ctx)

	connectionEventQueue := queue.NewInMemoryQueue()

	clientEventWorker := workers.NewClientEventWorker(workers.NewClientEventWorkerOptions{
		ClientManager:        clientManager,
		Repository:           repository,
		ConnectionEventQueue: connectionEventQueue,
	})
	go clientEventWorker.Start()

	stateManager := state.NewInMemoryStateManager()
	savePlayerStateChannelSize := 100
	savePlayerStateChan := make(chan workers.SavePlayerStateRequest, savePlayerStateChannelSize)

	saveLoopInterval := 10 * time.Second
	saveGameStateWorker := workers.NewSaveGameStateWorker(workers.NewSaveGameStateWorkerOptions{
		Repository:          repository,
		SavePlayerStateChan: savePlayerStateChan,
		StateManager:        stateManager,
		Interval:            saveLoopInterval,
	})
	go saveGameStateWorker.Start(ctx)

	gameLoopInterval := 100 * time.Millisecond // 10 FPS
	gameManager := game.NewGameManager(game.NewGameManagerOptions{
		ClientManager:        clientManager,
		ClientMessageQueue:   clientMessageQueue,
		ConnectionEventQueue: connectionEventQueue,
		Repository:           repository,
		StateManager:         stateManager,
		SavePlayerStateChan:  savePlayerStateChan,
		GameLoopInterval:     gameLoopInterval,
	})

	fmt.Println("Starting game manager")
	gameManager.Start(ctx)
}
