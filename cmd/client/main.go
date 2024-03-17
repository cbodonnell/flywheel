package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/cbodonnell/flywheel/pkg/game"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/solarlune/resolv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type GameMode int

const (
	GameModeMenu GameMode = iota
	GameModePlay
	GameModeOver
	GameModeNetworkError
)

func (m GameMode) String() string {
	switch m {
	case GameModeMenu:
		return "Menu"
	case GameModePlay:
		return "Play"
	case GameModeOver:
		return "Over"
	case GameModeNetworkError:
		return "NetworkError"
	}
	return "Unknown"
}

// Game implements ebiten.Game interface, which has Update, Draw and Layout methods.
type Game struct {
	// debug is a boolean value indicating whether debug mode is enabled.
	debug bool
	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// collisionSpace is the collision space.
	collisionSpace *resolv.Space
	// gameObjects is a map of game objects indexed by a unique identifier.
	gameObjects map[string]objects.GameObject
	// lastGameStateReceived is the timestamp of the last game state received from the server.
	lastGameStateReceived int64
	// gameStates is a buffer of game states received from the server.
	gameStates []*gametypes.GameState
	// mode is the current game mode.
	mode GameMode
	// touchIDs is the last touch identifiers.
	touchIDs []ebiten.TouchID
	// gamepadIDs is the last gamepad identifiers.
	gamepadIDs []ebiten.GamepadID
}

func NewGame(networkManager *network.NetworkManager) ebiten.Game {
	collisionSpace := game.NewCollisionSpace()

	return &Game{
		debug:          true,
		networkManager: networkManager,
		collisionSpace: collisionSpace,
		gameObjects:    make(map[string]objects.GameObject),
		mode:           GameModeMenu,
	}
}

var mplusNormalFont font.Face

func loadFonts() error {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}
	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %v", err)
	}

	return nil
}

// isKeyJustPressed returns a boolean value indicating whether the generic input is just pressed.
// This is used to handle both keyboard and touch inputs.
func (g *Game) isKeyJustPressed() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	if len(g.touchIDs) > 0 {
		return true
	}
	g.gamepadIDs = ebiten.AppendGamepadIDs(g.gamepadIDs[:0])
	for _, g := range g.gamepadIDs {
		if ebiten.IsStandardGamepadLayoutAvailable(g) {
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightBottom) {
				return true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(g, ebiten.StandardGamepadButtonRightRight) {
				return true
			}
		} else {
			// The button 0/1 might not be A/B buttons.
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton0) {
				return true
			}
			if inpututil.IsGamepadButtonJustPressed(g, ebiten.GamepadButton1) {
				return true
			}
		}
	}
	return false
}

func (g *Game) Update() error {
	switch g.mode {
	case GameModeMenu:
		if g.isKeyJustPressed() {
			if err := g.networkManager.Start(); err != nil {
				log.Error("Failed to start network manager: %v", err)
				g.networkManager.Stop()
				g.mode = GameModeNetworkError
				break
			}

			g.mode = GameModePlay
		}
	case GameModePlay:
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.networkManager.Stop()
			g.mode = GameModeOver
			g.gameObjects = make(map[string]objects.GameObject)
			break
		}

		if err := g.processPendingServerMessages(); err != nil {
			log.Error("Failed to process pending server messages: %v", err)
			break
		}

		if err := g.updatePlayerStates(); err != nil {
			log.Error("Failed to update player states: %v", err)
			break
		}

		if err := g.validateNetworkManager(); err != nil {
			log.Error("Network manager error: %v", err)
			g.networkManager.Stop()
			g.mode = GameModeNetworkError
			break
		}

		for _, obj := range g.gameObjects {
			if err := obj.Update(); err != nil {
				log.Error("Failed to update game object: %v", err)
				break
			}
		}
	case GameModeOver:
		if g.isKeyJustPressed() {
			g.mode = GameModeMenu
		}
	case GameModeNetworkError:
		if g.isKeyJustPressed() {
			g.mode = GameModeMenu
		}
	}

	return nil
}

func (g *Game) processPendingServerMessages() error {
	serverMessages, err := g.networkManager.ServerMessageQueue().ReadAllMessages()
	if err != nil {
		return fmt.Errorf("failed to read server messages: %v", err)
	}

	for _, item := range serverMessages {
		message, ok := item.(*messages.Message)
		if !ok {
			log.Error("Failed to cast message to messages.Message")
			continue
		}

		switch message.Type {
		case messages.MessageTypeServerGameUpdate:
			log.Trace("Received game state: %s", message.Payload)
			gameState := &gametypes.GameState{}
			if err := json.Unmarshal(message.Payload, gameState); err != nil {
				log.Error("Failed to unmarshal game state: %v", err)
				continue
			}

			if gameState.Timestamp < g.lastGameStateReceived {
				log.Warn("Received outdated game state: %d < %d", gameState.Timestamp, g.lastGameStateReceived)
				continue
			}
			g.lastGameStateReceived = gameState.Timestamp
			g.gameStates = append(g.gameStates, gameState)
		}
	}

	return nil
}

func (g *Game) updatePlayerStates() error {
	if len(g.gameStates) == 0 {
		return nil
	}

	// TODO: interpolate between game states
	gameState := g.gameStates[len(g.gameStates)-1]
	g.gameStates = g.gameStates[:0]

	for clientID, playerState := range gameState.Players {
		id := fmt.Sprintf("player-%d", clientID)
		if obj, ok := g.gameObjects[id]; !ok {
			log.Debug("Adding new player object for client %d", clientID)
			playerObject, err := objects.NewPlayer(id, g.networkManager, playerState)
			if err != nil {
				return fmt.Errorf("failed to create new player object: %v", err)
			}
			playerObject.State.Object = resolv.NewObject(playerState.Position.X, playerState.Position.Y, constants.PlayerWidth, constants.PlayerHeight, game.CollisionSpaceTagPlayer)
			g.collisionSpace.Add(playerObject.State.Object)
			g.gameObjects[id] = playerObject
		} else {
			playerObject, ok := obj.(*objects.Player)
			if !ok {
				return fmt.Errorf("failed to cast game object %s to *objects.Player", id)
			}
			playerObject.State.LastProcessedTimestamp = playerState.LastProcessedTimestamp
			playerObject.State.Position = playerState.Position
			playerObject.State.Velocity = playerState.Velocity
			playerObject.State.IsOnGround = playerState.IsOnGround
			playerObject.State.Object.Position.X = playerState.Position.X
			playerObject.State.Object.Position.Y = playerState.Position.Y
		}
	}

	return nil
}

func (g *Game) validateNetworkManager() error {
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
	g.drawOverlay(screen)
	for _, obj := range g.gameObjects {
		obj.Draw(screen)
	}
}

func (g *Game) drawOverlay(screen *ebiten.Image) {
	_, ping := g.networkManager.ServerTime()

	if g.debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.1f", ebiten.ActualFPS()))
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\nTPS: %0.1f", ebiten.ActualTPS()))
		ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\nPing: %0.1f", ping))
	}

	var t string
	switch g.mode {
	case GameModeMenu:
		t = "Press to Start!"
	case GameModePlay:
		return
	case GameModeOver:
		t = "Game Over"
	case GameModeNetworkError:
		t = "Network Error"
	}
	t = strings.ToUpper(t)
	bounds, _ := font.BoundString(mplusNormalFont, t)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(screen.Bounds().Dx())/2-float64(bounds.Max.X>>6)/2, float64(screen.Bounds().Dy())/2-float64(bounds.Max.Y>>6)/2)
	op.ColorScale.ScaleWithColor(color.White)
	text.DrawWithOptions(screen, t, mplusNormalFont, op)
}

const (
	ScreenWidth  = 640
	ScreenHeight = 480
)

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
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

	if err := loadFonts(); err != nil {
		panic(fmt.Sprintf("Failed to load fonts: %v", err))
	}

	serverMessageQueue := queue.NewInMemoryQueue(1024)
	networkManager, err := network.NewNetworkManager(serverMessageQueue)
	if err != nil {
		panic(fmt.Sprintf("Failed to create network manager: %v", err))
	}

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Flywheel Client")
	if err := ebiten.RunGame(NewGame(networkManager)); err != nil {
		panic(fmt.Sprintf("Failed to run game: %v", err))
	}
}
