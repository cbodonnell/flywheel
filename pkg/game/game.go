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
)

type GameManager struct {
	clientManager *clients.ClientManager
	messageQueue  *queue.MemoryQueue
	repository    repositories.Repository
	gameState     *types.GameState
	loopInterval  time.Duration
	stopChannel   chan struct{}
}

// NewGameManagerOptions contains options for creating a new GameManager.
type NewGameManagerOptions struct {
	ClientManager *clients.ClientManager
	MessageQueue  *queue.MemoryQueue
	Repository    repositories.Repository
	LoopInterval  time.Duration
}

func NewGameManager(opts NewGameManagerOptions) *GameManager {
	return &GameManager{
		clientManager: opts.ClientManager,
		messageQueue:  opts.MessageQueue,
		repository:    opts.Repository,
		loopInterval:  opts.LoopInterval,
		stopChannel:   make(chan struct{}),
	}
}

// StartGameLoop starts the game loop.
func (gm *GameManager) StartGameLoop(ctx context.Context) {
	// TODO: load player data when they connect
	gm.loadGameState(ctx)

	ticker := time.NewTicker(gm.loopInterval)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			gm.gameState.Timestamp = t.UnixMilli()
			gm.processMessages()
			gm.removeDisconnectedPlayers()
			// TODO: save non-critical data periodically, when the server shuts down, and when a player disconnects.
			// save critical data when it changes
			// TODO: save concurrently so it doesn't block the game loop
			gm.saveGameState(ctx)
			gm.broadcastGameState()

		case <-gm.stopChannel:
			fmt.Println("Game loop stopped.")
			return
		}
	}
}

// loadGameState loads the game state from the repository if it exists.
// Otherwise, it initializes an empty game state.
func (gm *GameManager) loadGameState(ctx context.Context) {
	lastGameState, err := gm.repository.LoadGameState(ctx)
	if err != nil {
		fmt.Printf("Error: failed to load game state: %v\n", err)
	}
	if lastGameState != nil {
		gm.gameState = lastGameState
	} else {
		fmt.Println("No game state found, initializing empty game state")
		gm.gameState = &types.GameState{
			Players: make(map[uint32]*types.PlayerState),
		}
	}
}

// StopGameLoop stops the game loop.
func (gm *GameManager) StopGameLoop() {
	close(gm.stopChannel)
}

// processMessages processes all pending messages in the queue
// and updates the game state accordingly.
func (gm *GameManager) processMessages() {
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
			gm.gameState.Players[message.ClientID] = clientPlayerUpdate.PlayerState
		default:
			fmt.Printf("Error: unhandled message type: %s\n", message.Type)
		}
	}
}

// removeDisconnectedPlayers removes disconnected clients from the game state.
func (gm *GameManager) removeDisconnectedPlayers() {
	for clientID := range gm.gameState.Players {
		if !gm.clientManager.Exists(clientID) {
			delete(gm.gameState.Players, clientID)
		}
	}
}

// saveGameState saves the game state to the repository.
func (gm *GameManager) saveGameState(ctx context.Context) {
	err := gm.repository.SaveGameState(ctx, gm.gameState)
	if err != nil {
		fmt.Printf("Error: failed to save game state: %v\n", err)
	}
}

// broadcastGameState sends the game state to connected clients.
func (gm *GameManager) broadcastGameState() {
	payload, err := json.Marshal(gm.gameState)
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
