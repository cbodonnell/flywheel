package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"
	"time"

	"github.com/cbodonnell/flywheel/client/flow"
	"github.com/cbodonnell/flywheel/client/fonts"
	"github.com/cbodonnell/flywheel/client/input"
	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/cbodonnell/flywheel/pkg/game"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/version"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"golang.org/x/image/font"
)

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
	// deletedObjects is a map of deleted game objects indexed by a unique identifier
	// and the timestamp of the deletion.
	deletedObjects map[string]int64
	// lastGameStateReceived is the timestamp of the last game state received from the server.
	lastGameStateReceived int64
	// gameStates is a buffer of game states received from the server.
	gameStates []*gametypes.GameState
	// mode is the current game mode.
	mode flow.GameMode
}

func NewGame(networkManager *network.NetworkManager) ebiten.Game {
	collisionSpace := game.NewCollisionSpace()

	return &Game{
		debug:          true,
		networkManager: networkManager,
		collisionSpace: collisionSpace,
		gameObjects:    make(map[string]objects.GameObject),
		deletedObjects: make(map[string]int64),
		mode:           flow.GameModeMenu,
	}
}

func (g *Game) Update() error {
	g.networkManager.UpdateServerTime(1.0 / float64(ebiten.TPS()))

	switch g.mode {
	case flow.GameModeMenu:
		if input.IsPositiveJustPressed() {
			if err := g.networkManager.Start(); err != nil {
				log.Error("Failed to start network manager: %v", err)
				g.networkManager.Stop()
				g.mode = flow.GameModeNetworkError
				break
			}

			g.mode = flow.GameModePlay
		}
	case flow.GameModePlay:
		if input.IsNegativeJustPressed() {
			g.networkManager.Stop()
			g.mode = flow.GameModeOver
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

		if err := g.checkNetworkManager(); err != nil {
			log.Error("Network manager error: %v", err)
			g.networkManager.Stop()
			g.mode = flow.GameModeNetworkError
			break
		}

		for _, obj := range g.gameObjects {
			if err := obj.Update(); err != nil {
				log.Error("Failed to update game object: %v", err)
				break
			}
		}

		if err := g.cleanupDeletedObjects(); err != nil {
			log.Error("Failed to cleanup deleted objects: %v", err)
			break
		}
	case flow.GameModeOver:
		if input.IsPositiveJustPressed() {
			g.mode = flow.GameModeMenu
		}
	case flow.GameModeNetworkError:
		if input.IsPositiveJustPressed() {
			g.mode = flow.GameModeMenu
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

			if err := g.reconcilePlayerState(gameState); err != nil {
				log.Warn("Failed to reconcile player state: %v", err)
			}
		case messages.MessageTypeServerPlayerConnect:
			playerConnect := &messages.ServerPlayerConnect{}
			if err := json.Unmarshal(message.Payload, playerConnect); err != nil {
				log.Error("Failed to unmarshal player connect message: %v", err)
				continue
			}

			id := fmt.Sprintf("player-%d", playerConnect.ClientID)
			if _, ok := g.gameObjects[id]; ok {
				log.Warn("Player object for client %d already exists", playerConnect.ClientID)
				continue
			}

			log.Debug("Adding new player object for client %d", playerConnect.ClientID)
			playerObject, err := objects.NewPlayer(id, g.networkManager, playerConnect.PlayerState)
			if err != nil {
				log.Error("Failed to create new player object: %v", err)
				continue
			}
			g.collisionSpace.Add(playerObject.State.Object)
			g.gameObjects[id] = playerObject
			delete(g.deletedObjects, id)
		case messages.MessageTypeServerPlayerDisconnect:
			log.Debug("Received player disconnect message: %s", message.Payload)
			playerDisconnect := &messages.ServerPlayerDisconnect{}
			if err := json.Unmarshal(message.Payload, playerDisconnect); err != nil {
				log.Error("Failed to unmarshal player disconnect message: %v", err)
				continue
			}

			id := fmt.Sprintf("player-%d", playerDisconnect.ClientID)
			obj, ok := g.gameObjects[id]
			if !ok {
				log.Warn("Player object for client %d not found", playerDisconnect.ClientID)
				continue
			}

			playerObject, ok := obj.(*objects.Player)
			if !ok {
				log.Error("Failed to cast game object %s to *objects.Player", id)
				continue
			}
			g.collisionSpace.Remove(playerObject.State.Object)
			delete(g.gameObjects, id)
			g.deletedObjects[id] = time.Now().UnixMilli()
		default:
			log.Warn("Received unexpected message type from server: %s", message.Type)
			continue
		}
	}

	return nil
}

func (g *Game) reconcilePlayerState(gameState *gametypes.GameState) error {
	playerState := gameState.Players[g.networkManager.ClientID()]
	if playerState == nil {
		return nil
	}

	playerObjectID := fmt.Sprintf("player-%d", g.networkManager.ClientID())
	obj, ok := g.gameObjects[playerObjectID]
	if !ok {
		log.Warn("Player object for client %d not found", g.networkManager.ClientID())
		return nil
	}

	playerObject, ok := obj.(*objects.Player)
	if !ok {
		return fmt.Errorf("failed to cast game object %s to *objects.Player", playerObjectID)
	}

	return playerObject.ReconcileState(playerState)
}

const (
	// InterpolationOffset is how far back in time we want to interpolate.
	// TODO: find a good rule of thumb for this value vs server tick rate
	InterpolationOffset = 150 // ms - currently 3x the server tick rate (50ms)
)

func (g *Game) updatePlayerStates() error {
	if len(g.gameStates) < 2 {
		return nil
	}

	serverTime, _ := g.networkManager.ServerTime()
	renderTime := int64(math.Round(serverTime)) - InterpolationOffset

	for len(g.gameStates) > 2 && g.gameStates[2].Timestamp < renderTime {
		g.gameStates = g.gameStates[1:]
	}

	if len(g.gameStates) > 2 {
		if err := g.interpolateState(g.gameStates[1], g.gameStates[2], renderTime); err != nil {
			return fmt.Errorf("failed to interpolate game state: %v", err)
		}
	} else {
		if err := g.extrapolateState(g.gameStates[0], g.gameStates[1], renderTime); err != nil {
			return fmt.Errorf("failed to extrapolate game state: %v", err)
		}
	}

	return nil
}

// interpolateState interpolates the game state between two states given a render time
// that is between the two states.
func (g *Game) interpolateState(from *gametypes.GameState, to *gametypes.GameState, renderTime int64) error {
	// we have a future state to interpolate to
	interpolationFactor := float64(renderTime-from.Timestamp) / float64(to.Timestamp-from.Timestamp)
	for clientID, playerState := range to.Players {
		if clientID == g.networkManager.ClientID() {
			continue
		}
		if _, ok := from.Players[clientID]; !ok {
			continue
		}
		previousPlayerState := from.Players[clientID]

		id := fmt.Sprintf("player-%d", clientID)
		if obj, ok := g.gameObjects[id]; !ok {
			if _, ok := g.deletedObjects[id]; ok {
				log.Warn("Player object for client %d was recently deleted, not instancing as part of update", clientID)
				continue
			}
			log.Debug("Adding new player object for client %d", clientID)
			playerObject, err := objects.NewPlayer(id, g.networkManager, playerState)
			if err != nil {
				return fmt.Errorf("failed to create new player object: %v", err)
			}
			g.collisionSpace.Add(playerObject.State.Object)
			g.gameObjects[id] = playerObject
		} else {
			playerObject, ok := obj.(*objects.Player)
			if !ok {
				return fmt.Errorf("failed to cast game object %s to *objects.Player", id)
			}
			playerObject.InterpolateState(previousPlayerState, playerState, interpolationFactor)
		}
	}

	return nil
}

// extrapolateState extrapolates the game state based on the last two states.
// This is used when we don't have a future state to interpolate to.
func (g *Game) extrapolateState(from *gametypes.GameState, to *gametypes.GameState, renderTime int64) error {
	// we don't have a future state, so we need to extrapolate from the last state
	extrapolationFactor := float64(renderTime-to.Timestamp) / float64(to.Timestamp-from.Timestamp)
	for clientID, playerState := range to.Players {
		if clientID == g.networkManager.ClientID() {
			continue
		}
		if _, ok := from.Players[clientID]; !ok {
			continue
		}
		previousPlayerState := from.Players[clientID]

		id := fmt.Sprintf("player-%d", clientID)
		if obj, ok := g.gameObjects[id]; !ok {
			log.Debug("Player object for client %d not found, not instancing since we're extrapolating", clientID)
		} else {
			playerObject, ok := obj.(*objects.Player)
			if !ok {
				return fmt.Errorf("failed to cast game object %s to *objects.Player", id)
			}
			playerObject.ExtrapolateState(previousPlayerState, playerState, extrapolationFactor)
		}
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

func (g *Game) cleanupDeletedObjects() error {
	for id, timestamp := range g.deletedObjects {
		if time.Now().UnixMilli()-timestamp > 2000 {
			delete(g.deletedObjects, id)
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, obj := range g.collisionSpace.Objects() {
		if obj.HasTags(game.CollisionSpaceTagLevel) {
			levelColor := color.RGBA{0x80, 0x80, 0x80, 0xff} // white
			vector.DrawFilledRect(screen, float32(obj.Position.X), float32(obj.Position.Y), float32(obj.Size.X), float32(obj.Size.Y), levelColor, false)
		}
	}

	playerObjectID := fmt.Sprintf("player-%d", g.networkManager.ClientID())
	for id, obj := range g.gameObjects {
		if id == playerObjectID {
			// skip the player object, we'll draw it last so it's on top
			continue
		}
		obj.Draw(screen)
	}
	if obj, ok := g.gameObjects[playerObjectID]; ok {
		obj.Draw(screen)
	}

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
