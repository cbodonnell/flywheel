package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/servers"
	"github.com/cbodonnell/flywheel/pkg/state"
)

type GameManager struct {
	clientManager    *clients.ClientManager
	messageQueue     *queue.InMemoryQueue
	repository       repositories.Repository
	stateManager     state.StateManager
	gameLoopInterval time.Duration
	saveLoopInterval time.Duration
}

// NewGameManagerOptions contains options for creating a new GameManager.
type NewGameManagerOptions struct {
	ClientManager    *clients.ClientManager
	MessageQueue     *queue.InMemoryQueue
	Repository       repositories.Repository
	StateManager     state.StateManager
	GameLoopInterval time.Duration
	SaveLoopInterval time.Duration
}

func NewGameManager(opts NewGameManagerOptions) *GameManager {
	return &GameManager{
		clientManager:    opts.ClientManager,
		messageQueue:     opts.MessageQueue,
		repository:       opts.Repository,
		stateManager:     opts.StateManager,
		gameLoopInterval: opts.GameLoopInterval,
		saveLoopInterval: opts.SaveLoopInterval,
	}
}

func (gm *GameManager) Start(ctx context.Context) {
	// Save loop runs in the background
	go gm.saveLoop(ctx)
	// Game loop runs in the foreground
	gm.gameLoop(ctx)
}

func (gm *GameManager) Stop() {
	// TODO: gracefully stop the game and save the game state
}

// gameLoop starts a loop that runs the game logic.
func (gm *GameManager) gameLoop(ctx context.Context) {
	ticker := time.NewTicker(gm.gameLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			err := gm.gameTick(ctx, t)
			if err != nil {
				fmt.Printf("Error: failed to do game loop: %v\n", err)
			}
		}
	}
}

// gameTick runs one iteration of the game loop.
func (gm *GameManager) gameTick(ctx context.Context, t time.Time) error {
	gameState, err := gm.stateManager.Get()
	if err != nil {
		return fmt.Errorf("failed to get current game state: %v", err)
	}

	gameState.Timestamp = t.UnixMilli()
	gm.addConnectedPlayers(ctx, gameState)
	gm.processMessages(gameState)
	gm.removeDisconnectedPlayers(ctx, gameState)
	gm.broadcastGameState(gameState)

	if err := gm.stateManager.Set(gameState); err != nil {
		return fmt.Errorf("failed to set game state: %v", err)
	}

	return nil
}

// addConnectedPlayers adds connected clients to the game state.
func (gm *GameManager) addConnectedPlayers(ctx context.Context, gameState *types.GameState) {
	for _, client := range gm.clientManager.GetClients() {
		if _, ok := gameState.Players[client.ID]; !ok {
			lastKnownState, err := gm.repository.LoadPlayerState(ctx, client.ID)
			if err == nil {
				gameState.Players[client.ID] = lastKnownState
			} else {
				fmt.Printf("Error: failed to get player state for client %d: %v\n", client.ID, err)
				fmt.Printf("Adding client %d to game state with default values\n", client.ID)
				gameState.Players[client.ID] = &types.PlayerState{
					P: types.Position{
						X: 0,
						Y: 0,
					},
				}
			}
		}
	}
}

// processMessages processes all pending messages in the queue
// and updates the game state accordingly.
func (gm *GameManager) processMessages(gameState *types.GameState) {
	pendingMessages := gm.messageQueue.ReadAllMessages()
	for _, item := range pendingMessages {
		message, ok := item.(*messages.Message)
		if !ok {
			fmt.Println("Error: failed to cast message to messages.Message")
			continue
		}

		switch message.Type {
		case messages.MessageTypeClientPlayerUpdate:
			clientPlayerUpdate := &messages.ClientPlayerUpdate{}
			err := json.Unmarshal(message.Payload, clientPlayerUpdate)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal player state: %v\n", err)
				continue
			}
			// TODO: validate the update before applying it
			gameState.Players[message.ClientID] = clientPlayerUpdate.PlayerState
		default:
			fmt.Printf("Error: unhandled message type: %s\n", message.Type)
		}
	}
}

// removeDisconnectedPlayers removes disconnected clients from the game state.
func (gm *GameManager) removeDisconnectedPlayers(ctx context.Context, gameState *types.GameState) {
	for clientID, playerState := range gameState.Players {
		if !gm.clientManager.Exists(clientID) {
			if err := gm.repository.SavePlayerState(ctx, gameState.Timestamp, clientID, playerState); err != nil {
				fmt.Printf("Error: failed to save player state for client %d: %v\n", clientID, err)
				// don't delete the player from the game state if we can't save their state
				continue
			}
			delete(gameState.Players, clientID)
		}
	}
}

// broadcastGameState sends the game state to connected clients.
func (gm *GameManager) broadcastGameState(gameState *types.GameState) {
	payload, err := json.Marshal(gameState)
	if err != nil {
		fmt.Printf("Error: failed to marshal game state: %v\n", err)
		return
	}

	for _, client := range gm.clientManager.GetClients() {
		message := &messages.Message{
			ClientID: 0, // ClientID 0 means the message is from the server
			Type:     messages.MessageTypeServerGameUpdate,
			Payload:  payload,
		}

		if client.UDPAddress == nil {
			// TODO: trace logging for stuff like this
			// fmt.Printf("Error: client %d does not have a UDP address\n", client.ID)
			continue
		}
		// TODO: reliable vs unreliable messages
		err := servers.WriteMessageToUDP(gm.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			fmt.Printf("Error: failed to write message to UDP connection for client %d: %v\n", client.ID, err)
		} else {
			fmt.Printf("Sent message: %s\n", message.Type)
		}
	}
}

// saveLoop starts a loop that periodically saves the game state.
func (gm *GameManager) saveLoop(ctx context.Context) {
	ticker := time.NewTicker(gm.saveLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			gameState, err := gm.stateManager.Get()
			if err != nil {
				fmt.Printf("Error: failed to get current game state: %v\n", err)
				continue
			}
			gameState.Timestamp = t.UnixMilli()
			gm.saveGameState(ctx, gameState)
		}
	}
}

// saveGameState saves the game state to the repository.
func (gm *GameManager) saveGameState(ctx context.Context, gameState *types.GameState) {
	err := gm.repository.SaveGameState(ctx, gameState)
	if err != nil {
		fmt.Printf("Error: failed to save game state: %v\n", err)
	}
}
