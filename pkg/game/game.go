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
	"github.com/cbodonnell/flywheel/pkg/workers"
)

type GameManager struct {
	clientManager        *clients.ClientManager
	clientMessageQueue   queue.Queue
	connectionEventQueue queue.Queue
	repository           repositories.Repository
	stateManager         state.StateManager
	savePlayerStateChan  chan<- workers.SavePlayerStateRequest
	gameLoopInterval     time.Duration
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
				fmt.Printf("Error: failed to do game loop: %v\n", err)
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
	pendingEvents := gm.connectionEventQueue.ReadAllMessages()
	for _, item := range pendingEvents {
		switch event := item.(type) {
		case *types.ConnectPlayerEvent:
			gameState.Players[event.ClientID] = event.PlayerState
		case *types.DisconnectPlayerEvent:
			// send a request to save the player state before deleting it
			saveRequest := workers.SavePlayerStateRequest{
				Timestamp:   gameState.Timestamp,
				ClientID:    event.ClientID,
				PlayerState: gameState.Players[event.ClientID],
			}
			gm.savePlayerStateChan <- saveRequest
			delete(gameState.Players, event.ClientID)
		default:
			fmt.Printf("Error: unhandled connection event type: %T\n", event)
		}
	}
}

// processClientMessages processes all pending client messages in the queue
// and updates the game state accordingly.
func (gm *GameManager) processClientMessages(gameState *types.GameState) {
	pendingMessages := gm.clientMessageQueue.ReadAllMessages()
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
			if _, ok := gameState.Players[message.ClientID]; !ok {
				fmt.Printf("Error: client %d does not have a player state\n", message.ClientID)
				continue
			}
			// TODO: validate the update before applying it
			gameState.Players[message.ClientID] = clientPlayerUpdate.PlayerState
		default:
			fmt.Printf("Error: unhandled message type: %s\n", message.Type)
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
			continue
		}

		// TODO: trace logging for stuff like this
		// fmt.Printf("Sent message: %s\n", message.Type)
	}
}
