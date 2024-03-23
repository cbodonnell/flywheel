package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/cbodonnell/flywheel/client/flow"
	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/client/input"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/scenes"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

// Game implements ebiten.Game interface, which has Update, Draw and Layout methods.
// TODO: move some of this into scenes
type Game struct {
	// debug is a boolean value indicating whether debug mode is enabled.
	debug bool
	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// mode is the current game mode.
	mode flow.GameMode

	scene scenes.Scene
}

func NewGame(networkManager *network.NetworkManager) (ebiten.Game, error) {
	g := &Game{
		debug:          true,
		networkManager: networkManager,
	}

	if err := g.loadMenu(); err != nil {
		return nil, fmt.Errorf("failed to load menu scene: %v", err)
	}

	return g, nil
}

func (g *Game) SetScene(scene scenes.Scene) error {
	if g.scene != nil {
		if err := g.scene.Destroy(); err != nil {
			return fmt.Errorf("failed to destroy previous scene: %v", err)
		}
	}

	g.scene = scene
	if err := g.scene.Init(); err != nil {
		return fmt.Errorf("failed to initialize scene: %v", err)
	}

	return nil
}

func (g *Game) loadMenu() error {
	menu, err := scenes.NewMenuScene()
	if err != nil {
		return fmt.Errorf("failed to create menu scene: %v", err)
	}
	if err := g.SetScene(menu); err != nil {
		return fmt.Errorf("failed to set menu scene: %v", err)
	}
	g.mode = flow.GameModeMenu
	return nil
}

func (g *Game) loadGame() error {
	gameScene, err := scenes.NewGameScene(g.networkManager)
	if err != nil {
		return fmt.Errorf("failed to create game scene: %v", err)
	}
	if err := g.SetScene(gameScene); err != nil {
		return fmt.Errorf("failed to set game scene: %v", err)
	}
	g.mode = flow.GameModePlay
	return nil
}

func (g *Game) loadGameOver() error {
	gameOver, err := scenes.NewGameOverScene()
	if err != nil {
		return fmt.Errorf("failed to create game over scene: %v", err)
	}
	if err := g.SetScene(gameOver); err != nil {
		return fmt.Errorf("failed to set game over scene: %v", err)
	}
	g.mode = flow.GameModeOver
	return nil
}

func (g *Game) loadNetworkError() error {
	networkError, err := scenes.NewMenuScene()
	if err != nil {
		return fmt.Errorf("failed to create network error scene: %v", err)
	}
	if err := g.SetScene(networkError); err != nil {
		return fmt.Errorf("failed to set network error scene: %v", err)
	}
	g.mode = flow.GameModeNetworkError
	return nil
}

func (g *Game) Update() error {
	g.networkManager.UpdateServerTime(1.0 / float64(ebiten.TPS()))

	switch g.mode {
	case flow.GameModeMenu:
		if input.IsPositiveJustPressed() {
			if err := g.networkManager.Start(); err != nil {
				log.Error("Failed to start network manager: %v", err)
				g.networkManager.Stop()
				if err := g.loadNetworkError(); err != nil {
					return fmt.Errorf("failed to load network error scene: %v", err)
				}
				break
			}

			if err := g.loadGame(); err != nil {
				return fmt.Errorf("failed to load game scene: %v", err)
			}
		}
	case flow.GameModePlay:
		if input.IsNegativeJustPressed() {
			g.networkManager.Stop()
			if err := g.loadGameOver(); err != nil {
				return fmt.Errorf("failed to load game over scene: %v", err)
			}
			break
		}

		if err := g.checkNetworkManager(); err != nil {
			log.Error("Network manager error: %v", err)
			g.networkManager.Stop()
			if err := g.loadNetworkError(); err != nil {
				return fmt.Errorf("failed to load network error scene: %v", err)
			}
		}
	case flow.GameModeOver:
		if input.IsPositiveJustPressed() {
			if err := g.loadMenu(); err != nil {
				return fmt.Errorf("failed to load menu scene: %v", err)
			}
		}
	case flow.GameModeNetworkError:
		if input.IsPositiveJustPressed() {
			if err := g.loadMenu(); err != nil {
				return fmt.Errorf("failed to load menu scene: %v", err)
			}
		}
	}

	if err := g.scene.Update(); err != nil {
		return fmt.Errorf("failed to update scene: %v", err)
	}

	return nil
}

// checkNetworkManager checks the network manager for errors and returns any that are found.
func (g *Game) checkNetworkManager() error {
	select {
	case err := <-g.networkManager.TCPClientErrChan():
		return fmt.Errorf("TCP client error: %v", err)
	case err := <-g.networkManager.UDPClientErrChan():
		return fmt.Errorf("UDP client error: %v", err)
	default:
		return nil
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
	g.drawOverlay(screen)
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	serverTime, ping := g.networkManager.ServerTime()

	if g.debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n   FPS: %0.1f", ebiten.ActualFPS()))
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n   TPS: %0.1f", ebiten.ActualTPS()))
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n   Ping: %0.1f", ping))
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n   Time: %0.0f", serverTime))
	}

	var t string
	switch g.mode {
	case flow.GameModeMenu:
		t = "Press to Start!"
	case flow.GameModePlay:
		return
	case flow.GameModeOver:
		t = "Game Over"
	case flow.GameModeNetworkError:
		t = "Network Error"
	}
	t = strings.ToUpper(t)
	bounds, _ := font.BoundString(fonts.MPlusNormalFont, t)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screen.Bounds().Dx())/2-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())/2-float64(bounds.Max.Y>>6)/2)
	op.ColorScale.ScaleWithColor(color.White)
	text.DrawWithOptions(screen, t, fonts.MPlusNormalFont, op)
}

const (
	DefaultScreenWidth  = 640
	DefaultScreenHeight = 480
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DefaultScreenWidth, DefaultScreenHeight
}

func main() {
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

	game, err := NewGame(networkManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to create game: %v", err))
	}

	ebiten.SetWindowSize(DefaultScreenWidth, DefaultScreenHeight)
	ebiten.SetWindowTitle("Flywheel Client")
	if err := ebiten.RunGame(game); err != nil {
		panic(fmt.Sprintf("Failed to run game: %v", err))
	}
}
