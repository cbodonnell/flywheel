package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cbodonnell/flywheel/client/game"
	clientgame "github.com/cbodonnell/flywheel/client/game"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	log.Info("Starting client version %s", version.Get())

	debug := flag.Bool("debug", false, "Debug mode")
	logLevel := flag.String("log-level", "info", "Log level")
	serverHostname := flag.String("server-hostname", network.DefaultServerHostname, "Server hostname")
	serverTCPPort := flag.Int("server-tcp-port", network.DefaultServerTCPPort, "Server TCP port")
	serverUDPPort := flag.Int("server-udp-port", network.DefaultServerUDPPort, "Server UDP port")
	authServerURL := flag.String("auth-server-url", game.DefaultAuthServerURL, "Auth server URL")
	apiServerURL := flag.String("api-server-url", game.DefaultAPIServerURL, "API server URL")
	automationEmail := flag.String("automation-email", "", "Automation email")
	automationPassword := flag.String("automation-password", "", "Automation password")
	flag.Parse()

	parsedLogLevel, err := log.ParseLogLevel(*logLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	logger := log.New(os.Stdout, "", log.DefaultLoggerFlag, parsedLogLevel)
	log.SetDefaultLogger(logger)
	log.Info("Log level set to %s", parsedLogLevel)

	serverSettings := network.ServerSettings{
		Hostname: *serverHostname,
		TCPPort:  *serverTCPPort,
		UDPPort:  *serverUDPPort,
	}
	serverMessageQueue := queue.NewInMemoryQueue(1024)
	networkManager, err := network.NewNetworkManager(serverSettings, serverMessageQueue)
	if err != nil {
		panic(fmt.Sprintf("Failed to create network manager: %v", err))
	}
	log.Info("Configured for game server %s ports %d (TCP) and %d (UDP)", serverSettings.Hostname, serverSettings.TCPPort, serverSettings.UDPPort)
	log.Info("Configured for auth server %s", *authServerURL)

	gameOpts := clientgame.NewGameOptions{
		Debug:          *debug,
		AuthURL:        *authServerURL,
		APIURL:         *apiServerURL,
		NetworkManager: networkManager,
	}
	if *automationEmail != "" && *automationPassword != "" {
		gameOpts.GameAutomation = &clientgame.GameAutomation{
			Email:    *automationEmail,
			Password: *automationPassword,
		}
	}
	game, err := clientgame.NewGame(gameOpts)
	if err != nil {
		panic(fmt.Sprintf("Failed to create game: %v", err))
	}

	ebiten.SetWindowSize(clientgame.DefaultScreenWidth, clientgame.DefaultScreenHeight)
	ebiten.SetWindowTitle("Flywheel Client")
	if err := ebiten.RunGame(game); err != nil {
		panic(fmt.Sprintf("Failed to run game: %v", err))
	}
}
