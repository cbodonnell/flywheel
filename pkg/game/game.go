package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/game/constants"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/servers"
	"github.com/cbodonnell/flywheel/pkg/state"
	"github.com/cbodonnell/flywheel/pkg/workers"
	"github.com/solarlune/resolv"
)

type GameManager struct {
	clientManager        *clients.ClientManager
	clientMessageQueue   queue.Queue
	connectionEventQueue queue.Queue
	repository           repositories.Repository
	stateManager         state.StateManager
	savePlayerStateChan  chan<- workers.SavePlayerStateRequest
	gameLoopInterval     time.Duration
	collisionSpace       *resolv.Space
}

// NewGameManagerOptions contains options for creating a new GameManager.
type NewGameManagerOptions struct {
	ClientManager        *clients.ClientManager
	ClientMessageQueue   queue.Queue
	ConnectionEventQueue queue.Queue
	Repository           repositories.Repository
	StateManager         state.StateManager
	SavePlayerStateChan  chan<- workers.SavePlayerStateRequest
	GameLoopInterval     time.Duration
	CollisionSpace       *resolv.Space
}

func NewGameManager(opts NewGameManagerOptions) *GameManager {
	return &GameManager{
		clientManager:        opts.ClientManager,
		clientMessageQueue:   opts.ClientMessageQueue,
		connectionEventQueue: opts.ConnectionEventQueue,
		repository:           opts.Repository,
		stateManager:         opts.StateManager,
		savePlayerStateChan:  opts.SavePlayerStateChan,
		gameLoopInterval:     opts.GameLoopInterval,
		collisionSpace:       opts.CollisionSpace,
	}
}

// Start starts the game loop.
func (gm *GameManager) Start(ctx context.Context) {
	ticker := time.NewTicker(gm.gameLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			err := gm.gameTick(ctx, t)
			if err != nil {
				log.Error("Failed to do game loop: %v", err)
			}
		}
	}
}

func (gm *GameManager) Stop() {
	// TODO: gracefully stop the game and save the game state
}

// gameTick runs one iteration of the game loop.
func (gm *GameManager) gameTick(ctx context.Context, t time.Time) error {
	gameState, err := gm.stateManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get current game state: %v", err)
	}

	gameState.Timestamp = t.UnixMilli()
	gm.processConnectionEvents(gameState)
	gm.processClientMessages(gameState)
	gm.broadcastGameState(gameState)

	if err := gm.stateManager.Set(gameState); err != nil {
		return fmt.Errorf("failed to set game state: %v", err)
	}

	return nil
}

// processConnectionEvents processes all pending connection events in the queue
// and updates the game state accordingly.
func (gm *GameManager) processConnectionEvents(gameState *types.GameState) {
	pendingEvents, err := gm.connectionEventQueue.ReadAllMessages()
	if err != nil {
		log.Error("Failed to read connection events: %v", err)
		return
	}
	for _, item := range pendingEvents {
		switch event := item.(type) {
		case *types.ConnectPlayerEvent:
			playerState := event.PlayerState
			playerState.Object = resolv.NewObject(playerState.Position.X, playerState.Position.Y, constants.PlayerWidth, constants.PlayerHeight)
			// add the player to the game state
			gameState.Players[event.ClientID] = playerState
			// add the player object to the collision space
			gm.collisionSpace.Add(playerState.Object)
		case *types.DisconnectPlayerEvent:
			// send a request to save the player state before deleting it
			saveRequest := workers.SavePlayerStateRequest{
				Timestamp:   gameState.Timestamp,
				ClientID:    event.ClientID,
				PlayerState: gameState.Players[event.ClientID],
			}
			gm.savePlayerStateChan <- saveRequest
			// remove the player object from the collision space
			gm.collisionSpace.Remove(gameState.Players[event.ClientID].Object)
			// delete the player from the game state
			delete(gameState.Players, event.ClientID)
		default:
			log.Error("unhandled connection event type: %T", event)
		}
	}
}

// processClientMessages processes all pending client messages in the queue
// and updates the game state accordingly.
func (gm *GameManager) processClientMessages(gameState *types.GameState) {
	pendingMessages, err := gm.clientMessageQueue.ReadAllMessages()
	if err != nil {
		log.Error("Failed to read client messages: %v", err)
		return
	}
	for _, item := range pendingMessages {
		message, ok := item.(*messages.Message)
		if !ok {
			log.Error("Failed to cast message to messages.Message")
			continue
		}

		switch message.Type {
		case messages.MessageTypeClientPlayerUpdate:
			clientPlayerUpdate := &messages.ClientPlayerUpdate{}
			err := json.Unmarshal(message.Payload, clientPlayerUpdate)
			if err != nil {
				log.Error("Failed to unmarshal player state: %v", err)
				continue
			}
			if _, ok := gameState.Players[message.ClientID]; !ok {
				log.Warn("Client %d is not in the game state", message.ClientID)
				continue
			}

			if gameState.Players[message.ClientID].LastProcessedTimestamp > clientPlayerUpdate.Timestamp {
				log.Warn("Client %d sent an outdated player update", message.ClientID)
				continue
			}

			// TODO: validate the update before applying it
			updatePlayerState(gameState.Players[message.ClientID], clientPlayerUpdate)
		default:
			log.Error("Unhandled message type: %s", message.Type)
		}
	}
}

// updatePlayerState updates the player's position and velocity based on the
// client's input and the game state.
// The player state is updated in place.
func updatePlayerState(playerState *types.PlayerState, clientPlayerUpdate *messages.ClientPlayerUpdate) {
	// X-axis
	// Apply input
	dx := kinematic.Displacement(clientPlayerUpdate.InputX*constants.PlayerSpeed, clientPlayerUpdate.DeltaTime, 0)
	vx := kinematic.FinalVelocity(clientPlayerUpdate.InputX*constants.PlayerSpeed, clientPlayerUpdate.DeltaTime, 0)

	// Check for collisions
	if collision := playerState.Object.Check(dx, 0); collision != nil {
		dx = collision.ContactWithObject(collision.Objects[0]).X
		vx = 0
	}

	// Update player state
	playerState.Position.X += dx
	playerState.Velocity.X = vx

	// Y-axis
	// Apply gravity if not on ground
	dy := kinematic.Displacement(playerState.Velocity.Y, clientPlayerUpdate.DeltaTime, kinematic.Gravity)
	vy := kinematic.FinalVelocity(playerState.Velocity.Y, clientPlayerUpdate.DeltaTime, kinematic.Gravity)

	// Check for collisions
	if collision := playerState.Object.Check(0, dy); collision != nil {
		dy = collision.ContactWithObject(collision.Objects[0]).Y
		vy = 0
	}

	// Update player state
	playerState.Position.Y += dy
	playerState.Velocity.Y = vy
	playerState.LastProcessedTimestamp = clientPlayerUpdate.Timestamp
}

// broadcastGameState sends the game state to connected clients.
func (gm *GameManager) broadcastGameState(gameState *types.GameState) {
	payload, err := json.Marshal(gameState)
	if err != nil {
		log.Error("Failed to marshal game state: %v", err)
		return
	}

	for _, client := range gm.clientManager.GetClients() {
		message := &messages.Message{
			ClientID: 0, // ClientID 0 means the message is from the server
			Type:     messages.MessageTypeServerGameUpdate,
			Payload:  payload,
		}

		if client.UDPAddress == nil {
			log.Trace("Client %d does not have a UDP address", client.ID)
			continue
		}
		// TODO: reliable vs unreliable messages
		err := servers.WriteMessageToUDP(gm.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			log.Error("Failed to write message to UDP connection for client %d: %v", client.ID, err)
			continue
		}

		log.Trace("Sent message: %s", message.Type)
	}
}
