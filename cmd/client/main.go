package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"github.com/cbodonnell/flywheel/client/network"
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
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// collisionSpace is the collision space.
	collisionSpace *resolv.Space
	// playerStates is a map of the current player states indexed by client ID.
	playerStates map[uint32]*gametypes.PlayerState
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
		networkManager: networkManager,
		collisionSpace: collisionSpace,
		playerStates:   make(map[uint32]*gametypes.PlayerState),
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
				return nil
			}

			g.mode = GameModePlay
		}
	case GameModePlay:
		if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.networkManager.Stop()
			g.mode = GameModeOver
			g.playerStates = make(map[uint32]*gametypes.PlayerState)
			break
		}

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

				for clientID, playerState := range gameState.Players {
					if _, ok := g.playerStates[clientID]; !ok {
						log.Debug("Adding new player state for client %d", clientID)
						g.playerStates[clientID] = playerState
						g.playerStates[clientID].Object = resolv.NewObject(float64(playerState.Position.X), float64(playerState.Position.Y), constants.PlayerWidth, constants.PlayerHeight)
						g.collisionSpace.Add(g.playerStates[clientID].Object)
					} else {
						g.playerStates[clientID].Position = playerState.Position
						g.playerStates[clientID].Velocity = playerState.Velocity
						g.playerStates[clientID].IsOnGround = playerState.IsOnGround
						g.playerStates[clientID].Object.Position.X = float64(playerState.Position.X)
						g.playerStates[clientID].Object.Position.Y = float64(playerState.Position.Y)
					}
				}
			}
		}

		if err := g.validateNetworkManager(); err != nil {
			log.Error("Network manager error: %v", err)
			g.networkManager.Stop()
			g.mode = GameModeNetworkError
			break
		}

		localPlayer := g.playerStates[g.networkManager.ClientID()]
		if localPlayer == nil {
			log.Warn("Local player state not found")
			break
		}

		inputX := 0.0
		rightPressed := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
		leftPressed := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
		if rightPressed && !leftPressed {
			inputX = 1.0
		} else if leftPressed && !rightPressed {
			inputX = -1.0
		}

		inputY := 0.0
		upPressed := ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
		downPressed := ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)
		if upPressed && !downPressed {
			inputY = -1.0
		} else if downPressed && !upPressed {
			inputY = 1.0
		}

		inputJump := ebiten.IsKeyPressed(ebiten.KeySpace)

		cpu := &messages.ClientPlayerUpdate{
			Timestamp: time.Now().UnixMilli(),
			InputX:    inputX,
			InputY:    inputY,
			InputJump: inputJump,
			DeltaTime: 1.0 / 60.0,
		}
		payload, err := json.Marshal(cpu)
		if err != nil {
			return fmt.Errorf("failed to marshal client player update: %v", err)
		}

		msg := &messages.Message{
			ClientID: g.networkManager.ClientID(),
			Type:     messages.MessageTypeClientPlayerUpdate,
			Payload:  payload,
		}

		if err := g.networkManager.SendUnreliableMessage(msg); err != nil {
			return fmt.Errorf("failed to send client player update: %v", err)
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
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\nTPS: %0.2f", ebiten.ActualTPS()))
	var t string
	switch g.mode {
	case GameModeMenu:
		t = "Press to Start!"
	case GameModePlay:
		for _, playerState := range g.playerStates {
			playerObject := playerState.Object
			playerColor := color.RGBA{0, 255, 60, 255} // Green
			if playerState.IsOnGround {
				playerColor = color.RGBA{200, 0, 200, 255} // Purple
			}
			vector.DrawFilledRect(screen, float32(playerObject.Position.X), float32(ScreenHeight-constants.PlayerHeight)-float32(playerObject.Position.Y), float32(playerObject.Size.X), float32(playerObject.Size.Y), playerColor, false)
		}
	case GameModeOver:
		t = "Game Over"
	case GameModeNetworkError:
		t = "Network Error"
	}
	t = strings.ToUpper(t)
	bounds, _ := font.BoundString(mplusNormalFont, t)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ScreenWidth/2-float64(bounds.Max.X>>6)/2, ScreenHeight/2-float64(bounds.Max.Y>>6)/2)
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

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Flywheel Client")
	if err := ebiten.RunGame(NewGame(networkManager)); err != nil {
		panic(fmt.Sprintf("Failed to run game: %v", err))
	}
}
