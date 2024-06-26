package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/cbodonnell/flywheel/pkg/api"
	"github.com/cbodonnell/flywheel/pkg/auth"
	authhandlers "github.com/cbodonnell/flywheel/pkg/auth/handlers"
	authproviders "github.com/cbodonnell/flywheel/pkg/auth/providers"
	"github.com/cbodonnell/flywheel/pkg/game"
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
	wsPort := flag.Int("ws-port", 8890, "WebSocket port to listen on")
	wsAllowOrigin := flag.String("ws-allow-origin", "http://localhost:3000", "comma-separated list of allowed origins for the websocket server")
	authPort := flag.Int("auth-port", 8080, "Auth server port")
	authAllowOrigin := flag.String("auth-allow-origin", "http://localhost:3000", "comma-separated list of allowed origins for the auth server")
	apiPort := flag.Int("api-port", 9090, "API server port")
	apiAllowOrigin := flag.String("api-allow-origin", "http://localhost:3000", "comma-separated list of allowed origins for the api server")
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

	firebaseApiKey := os.Getenv("FLYWHEEL_FIREBASE_API_KEY")
	if firebaseApiKey == "" {
		panic("FLYWHEEL_FIREBASE_API_KEY environment variable must be set")
	}
	authServerOpts := auth.NewAuthServerOptions{
		Port:        *authPort,
		AllowOrigin: *authAllowOrigin,
		Handler:     authhandlers.NewFirebaseAuthHandler(firebaseApiKey),
	}
	authTLSCertFile := os.Getenv("FLYWHEEL_AUTH_TLS_CERT_FILE")
	authTLSKeyFile := os.Getenv("FLYWHEEL_AUTH_TLS_KEY_FILE")
	if authTLSCertFile != "" && authTLSKeyFile != "" {
		authServerOpts.TLS = &auth.TLSConfig{
			CertFile: authTLSCertFile,
			KeyFile:  authTLSKeyFile,
		}
	}
	authServer := auth.NewAuthServer(authServerOpts)
	go authServer.Start()

	connectionEventChan := make(chan network.ConnectionEvent, 100)
	clientManager := network.NewClientManager(connectionEventChan)

	firebaseProjectID := os.Getenv("FLYWHEEL_FIREBASE_PROJECT_ID")
	if firebaseProjectID == "" {
		panic("FLYWHEEL_FIREBASE_PROJECT_ID environment variable must be set")
	}
	authProvider, err := authproviders.NewFirebaseAuthProvider(ctx, firebaseProjectID)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Firebase auth provider: %v", err))
	}

	clientMessageQueue := queue.NewInMemoryQueue(10000)
	networkManagerOpts := network.NewNetworkManagerOptions{
		AuthProvider:  authProvider,
		ClientManager: clientManager,
		MessageQueue:  clientMessageQueue,
		TCPPort:       *tcpPort,
		UDPPort:       *udpPort,
		WSPort:        *wsPort,
		WSAllowOrigin: *wsAllowOrigin,
	}
	wsTLSCertFile := os.Getenv("FLYWHEEL_WS_TLS_CERT_FILE")
	wsTLSKeyFile := os.Getenv("FLYWHEEL_WS_TLS_KEY_FILE")
	if wsTLSCertFile != "" && wsTLSKeyFile != "" {
		networkManagerOpts.WSTLS = &network.TLSConfig{
			CertFile: wsTLSCertFile,
			KeyFile:  wsTLSKeyFile,
		}
	}
	networkManager := network.NewNetworkManager(networkManagerOpts)
	go networkManager.Start(ctx)

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

	apiServerOpts := api.NewAPIServerOptions{
		Port:         *apiPort,
		AllowOrigin:  *apiAllowOrigin,
		AuthProvider: authProvider,
		Repository:   repository,
	}
	apiTLSCertFile := os.Getenv("FLYWHEEL_API_TLS_CERT_FILE")
	apiTLSKeyFile := os.Getenv("FLYWHEEL_API_TLS_KEY_FILE")
	if apiTLSCertFile != "" && apiTLSKeyFile != "" {
		apiServerOpts.TLS = &api.TLSConfig{
			CertFile: apiTLSCertFile,
			KeyFile:  apiTLSKeyFile,
		}
	}
	apiServer := api.NewAPIServer(apiServerOpts)
	go apiServer.Start()

	serverEventQueue := queue.NewInMemoryQueue(1000)
	connectionEventWorker := workers.NewConnectionEventWorker(workers.NewConnectionEventWorkerOptions{
		ConnectionEventChan: connectionEventChan,
		Repository:          repository,
		ServerEventQueue:    serverEventQueue,
	})
	go connectionEventWorker.Start(ctx)

	saveStateChan := make(chan workers.SaveStateRequest, 100)
	saveGameStateWorker := workers.NewSaveGameStateWorker(workers.NewSaveGameStateWorkerOptions{
		Repository:    repository,
		SaveStateChan: saveStateChan,
	})
	go saveGameStateWorker.Start(ctx)

	broadcastMessageChan := make(chan workers.BroadcastMessage, 100)
	broadcastMessageWorker := workers.NewBroadcastMessageWorker(workers.NewBroadcastMessageWorkerOptions{
		NetworkManager:       networkManager,
		BroadcastMessageChan: broadcastMessageChan,
	})
	go broadcastMessageWorker.Start(ctx)

	gameManager := game.NewGameManager(game.NewGameManagerOptions{
		ClientMessageQueue:   clientMessageQueue,
		ServerEventQueue:     serverEventQueue,
		SaveStateChan:        saveStateChan,
		BroadcastMessageChan: broadcastMessageChan,
		GameLoopInterval:     50 * time.Millisecond, // 20 ticks per second
		SaveStateInterval:    5 * time.Second,
	})

	log.Info("Starting game manager")
	if err := gameManager.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start game manager: %v", err))
	}
}
