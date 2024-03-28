package scenes

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/cbodonnell/flywheel/client/network"
	"github.com/cbodonnell/flywheel/client/objects"
	"github.com/cbodonnell/flywheel/pkg/game"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
)

type GameScene struct {
	BaseScene

	// networkManager is the network manager.
	networkManager *network.NetworkManager
	// collisionSpace is the collision space.
	collisionSpace *resolv.Space
	// deletedObjects is a map of deleted game objects indexed by a unique identifier
	// and the timestamp of the deletion.
	deletedObjects map[string]int64
	// lastGameStateReceived is the timestamp of the last game state received from the server.
	lastGameStateReceived int64
	// gameStates is a buffer of game states received from the server.
	gameStates []*gametypes.GameState
}

func NewGameScene(networkManager *network.NetworkManager) (Scene, error) {
	return &GameScene{
		BaseScene: BaseScene{
			Root: objects.NewBaseObject(),
		},
		networkManager: networkManager,
		collisionSpace: game.NewCollisionSpace(),
		deletedObjects: make(map[string]int64),
	}, nil
}

func (g *GameScene) Update() error {
	if err := g.processPendingServerMessages(); err != nil {
		return fmt.Errorf("failed to process pending server messages: %v", err)
	}

	if err := g.updateGameState(); err != nil {
		return fmt.Errorf("failed to update game state: %v", err)
	}

	if err := objects.UpdateTree(g.Root); err != nil {
		return fmt.Errorf("failed to update object tree: %v", err)
	}

	if err := g.cleanupDeletedObjects(); err != nil {
		return fmt.Errorf("failed to cleanup deleted objects: %v", err)
	}

	return nil
}

func (g *GameScene) processPendingServerMessages() error {
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
			obj := g.GetRoot().GetChild(id)
			if obj != nil {
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
			g.GetRoot().AddChild(id, playerObject)
			delete(g.deletedObjects, id)
		case messages.MessageTypeServerPlayerDisconnect:
			playerDisconnect := &messages.ServerPlayerDisconnect{}
			if err := json.Unmarshal(message.Payload, playerDisconnect); err != nil {
				log.Error("Failed to unmarshal player disconnect message: %v", err)
				continue
			}

			id := fmt.Sprintf("player-%d", playerDisconnect.ClientID)
			obj := g.GetRoot().GetChild(id)
			if obj == nil {
				log.Warn("Player object for client %d not found", playerDisconnect.ClientID)
				continue
			}
			playerObject, ok := obj.(*objects.Player)
			if !ok {
				log.Error("Failed to cast game object %s to *objects.Player", id)
				continue
			}
			log.Debug("Removing player object for client %d", playerDisconnect.ClientID)
			g.collisionSpace.Remove(playerObject.State.Object)
			if err := g.GetRoot().RemoveChild(id); err != nil {
				log.Error("Failed to remove player object: %v", err)
				continue
			}
			g.deletedObjects[id] = time.Now().UnixMilli()
		default:
			log.Warn("Received unexpected message type from server: %s", message.Type)
			continue
		}
	}

	return nil
}

func (g *GameScene) reconcilePlayerState(gameState *gametypes.GameState) error {
	playerState := gameState.Players[g.networkManager.ClientID()]
	if playerState == nil {
		return nil
	}

	playerObjectID := fmt.Sprintf("player-%d", g.networkManager.ClientID())
	obj := g.GetRoot().GetChild(playerObjectID)
	if obj == nil {
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

func (g *GameScene) updateGameState() error {
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
func (g *GameScene) interpolateState(from *gametypes.GameState, to *gametypes.GameState, renderTime int64) error {
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
		obj := g.GetRoot().GetChild(id)
		if obj == nil {
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
			if err := g.GetRoot().AddChild(id, playerObject); err != nil {
				return fmt.Errorf("failed to add player object: %v", err)
			}
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
func (g *GameScene) extrapolateState(from *gametypes.GameState, to *gametypes.GameState, renderTime int64) error {
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
		obj := g.GetRoot().GetChild(id)
		if obj == nil {
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

func (g *GameScene) cleanupDeletedObjects() error {
	for id, timestamp := range g.deletedObjects {
		if time.Now().UnixMilli()-timestamp > 2000 {
			delete(g.deletedObjects, id)
		}
	}
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	for _, obj := range g.collisionSpace.Objects() {
		if obj.HasTags(game.CollisionSpaceTagLevel) {
			levelColor := color.RGBA{0x80, 0x80, 0x80, 0xff} // white
			vector.DrawFilledRect(screen, float32(obj.Position.X), float32(obj.Position.Y), float32(obj.Size.X), float32(obj.Size.Y), levelColor, false)
		}
	}

	playerObjectID := fmt.Sprintf("player-%d", g.networkManager.ClientID())
	for id, obj := range g.GetRoot().GetChildren() {
		if id == playerObjectID {
			// skip the player object, we'll draw it last so it's on top
			continue
		}
		objects.DrawTree(obj, screen)
	}
	obj := g.GetRoot().GetChild(playerObjectID)
	if obj != nil {
		objects.DrawTree(obj, screen)
	}
}