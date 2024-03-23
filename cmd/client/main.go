package main

import (
	"flag"
	"fmt"
	"os"

	clientgame "github.com/cbodonnell/flywheel/client/game"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	debug := flag.Bool("debug", false, "Debug mode")
	logLevel := flag.String("log-level", "info", "Log level")
	flag.Parse()

	parsedLogLevel, err := log.ParseLogLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	logger := log.New(os.Stdout, "", log.DefaultLoggerFlag, parsedLogLevel)
	log.SetDefaultLogger(logger)
	log.Info("Log level set to %s", parsedLogLevel)

	log.Info("Starting client version %s", version.Get())

	serverMessageQueue := queue.NewInMemoryQueue(1024)
	networkManager, err := network.NewNetworkManager(serverMessageQueue)
	if err != nil {
		panic(fmt.Sprintf("Failed to create network manager: %v", err))
	}

	game, err := clientgame.NewGame(clientgame.NewGameOptions{
		Debug:          *debug,
		NetworkManager: networkManager,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create game: %v", err))
	}

	ebiten.SetWindowSize(clientgame.DefaultScreenWidth, clientgame.DefaultScreenHeight)
	ebiten.SetWindowTitle("Flywheel Client")
	if err := ebiten.RunGame(game); err != nil {
		panic(fmt.Sprintf("Failed to run game: %v", err))
	}
}
