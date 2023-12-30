package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/servers"
	"github.com/cbodonnell/flywheel/pkg/state"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/cbodonnell/flywheel/pkg/workers"
)

func main() {
	tcpPort := flag.String("tcp-port", "8888", "TCP port to listen on")
	udpPort := flag.String("udp-port", "8889", "UDP port to listen on")
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	parsedLogLevel, err := log.ParseLogLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	logger := log.New(os.Stdout, "", log.DefaultLoggerFlag, parsedLogLevel)
	log.SetDefaultLogger(logger)
	log.Info("Log level set to %s", parsedLogLevel)

	log.Info("Starting server version %s", version.Get())
	ctx := context.Background()

	clientManager := clients.NewClientManager()
	clientMessageQueue := queue.NewInMemoryQueue(10000)

	tcpServer := servers.NewTCPServer(clientManager, clientMessageQueue, *tcpPort)
	udpServer := servers.NewUDPServer(clientManager, clientMessageQueue, *udpPort)
	go tcpServer.Start()
	go udpServer.Start()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		panic("DATABASE_URL environment variable must be set")
	}
	repository := repositories.NewPostgresRepository(ctx, connStr)
	defer repository.Close(ctx)

	connectionEventQueue := queue.NewInMemoryQueue(1000)

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

	log.Info("Starting game manager")
	gameManager.Start(ctx)
}
