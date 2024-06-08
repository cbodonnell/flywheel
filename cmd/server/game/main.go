package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/network"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/cbodonnell/flywheel/pkg/workers"
)

func main() {
	tcpPort := flag.Int("tcp-port", 8888, "TCP port to listen on")
	udpPort := flag.Int("udp-port", 8889, "UDP port to listen on")
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	parsedLogLevel, err := log.ParseLogLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	logger := log.New(os.Stdout, "", log.DefaultLoggerFlag, parsedLogLevel)
	log.SetDefaultLogger(logger)
	log.Info("Log level set to %s", parsedLogLevel)

	log.Info("Starting game server version %s", version.Get())
	ctx := context.Background()

	clientManager := network.NewClientManager()
	clientMessageQueue := queue.NewInMemoryQueue(10000)

	firebaseProjectID := os.Getenv("FLYWHEEL_FIREBASE_PROJECT_ID")
	if firebaseProjectID == "" {
		panic("FLYWHEEL_FIREBASE_PROJECT_ID environment variable must be set")
	}
	authProvider, err := authproviders.NewFirebaseAuthProvider(ctx, firebaseProjectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Firebase auth provider: %v", err))
	}

	tcpServer := network.NewTCPServer(authProvider, clientManager, clientMessageQueue, *tcpPort)
	udpServer := network.NewUDPServer(clientManager, clientMessageQueue, *udpPort)
	go tcpServer.Start()
	go udpServer.Start()

	connStr := os.Getenv("FLYWHEEL_DATABASE_URL")
	if connStr == "" {
		connStr = "sqlite://flywheel.db"
	}

	u, err := url.Parse(connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse connection string: %v", err))
	}

	var repository repositories.Repository
	switch u.Scheme {
	case "sqlite":
		repository, err = repositories.NewSQLiteRepository(ctx, u.Host, "./migrations/sqlite")
		if err != nil {
			panic(fmt.Sprintf("Failed to create SQLite repository: %v", err))
		}
	case "postgresql":
		repository, err = repositories.NewPostgresRepository(ctx, u.String())
		if err != nil {
			panic(fmt.Sprintf("Failed to create Postgres repository: %v", err))
		}
	default:
		panic(fmt.Sprintf("Unknown database type %s", u.Scheme))
	}
	defer repository.Close(ctx)

	serverEventQueue := queue.NewInMemoryQueue(1000)

	connectionEventWorker := workers.NewConnectionEventWorker(workers.NewConnectionEventWorkerOptions{
		ClientManager:    clientManager,
		Repository:       repository,
		ServerEventQueue: serverEventQueue,
	})
	go connectionEventWorker.Start()

	savePlayerStateChannelSize := 100
	savePlayerStateChan := make(chan workers.SavePlayerStateRequest, savePlayerStateChannelSize)

	gameState := types.NewGameState(game.NewCollisionSpace())
	saveLoopInterval := 10 * time.Second
	saveGameStateWorker := workers.NewSaveGameStateWorker(workers.NewSaveGameStateWorkerOptions{
		Repository:          repository,
		SavePlayerStateChan: savePlayerStateChan,
		GameState:           gameState, // TODO: only the game loop should have access to this
		Interval:            saveLoopInterval,
	})
	go saveGameStateWorker.Start(ctx)

	serverMessageChannelSize := 100
	serverMessageChan := make(chan workers.ServerMessage, serverMessageChannelSize)

	serverMessageWorker := workers.NewServerMessageWorker(workers.NewServerMessageWorkerOptions{
		ClientManager:     clientManager,
		ServerMessageChan: serverMessageChan,
	})
	go serverMessageWorker.Start()

	gameLoopInterval := 50 * time.Millisecond // 20 ticks per second
	gameManager := game.NewGameManager(game.NewGameManagerOptions{
		ClientMessageQueue:  clientMessageQueue,
		ServerEventQueue:    serverEventQueue,
		Repository:          repository,
		GameState:           gameState,
		SavePlayerStateChan: savePlayerStateChan,
		ServerMessageChan:   serverMessageChan,
		GameLoopInterval:    gameLoopInterval,
	})

	log.Info("Starting game manager")
	if err := gameManager.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start game manager: %v", err))
	}
}
