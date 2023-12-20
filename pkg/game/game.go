package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/servers"
)

type GameState struct {
	// Timestamp is the time at which the game state was generated
	Timestamp int64 `json:"timestamp"`
	// Players maps client IDs to player states
	Players map[uint32]*PlayerState `json:"players"`
}

type PlayerState struct {
	P Position `json:"p"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ClientPlayerInput struct {
    Horizontal float64 `json:"horizontal"`
    Vertical   float64 `json:"vertical"`
    Jump       bool    `json:"jump"`
}

type GameManager struct {
	clientManager *clients.ClientManager
	messageQueue  *queue.MemoryQueue
	gameState     *GameState
	loopInterval  time.Duration
	stopChannel   chan struct{}
}

// NewGameManager creates a new game manager.
func NewGameManager(clientManager *clients.ClientManager, messageQueue *queue.MemoryQueue, loopInterval time.Duration) *GameManager {
	return &GameManager{
		clientManager: clientManager,
		messageQueue:  messageQueue,
		gameState: &GameState{
			Players: make(map[uint32]*PlayerState),
		},
		loopInterval: loopInterval,
		stopChannel:  make(chan struct{}),
	}
}

// StartGameLoop starts the game loop.
func (gm *GameManager) StartGameLoop() {
	ticker := time.NewTicker(gm.loopInterval)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			timestamp := t.UnixMilli()
			gm.processMessages(timestamp)
			gm.broadcastGameState()

		case <-gm.stopChannel:
			fmt.Println("Game loop stopped.")
			return
		}
	}
}

// StopGameLoop stops the game loop.
func (gm *GameManager) StopGameLoop() {
	close(gm.stopChannel)
}

func (gm *GameManager) processMessages(timestamp int64) {
	pendingMessages := gm.messageQueue.ReadAllMessages()
	for _, item := range pendingMessages {
		message, ok := item.(*messages.Message)
		if !ok {
			fmt.Println("Error: failed to cast message to messages.Message")
			continue
		}
		fmt.Printf("Received message: %+v\n", message)

		switch message.Type {

		case messages.MessageTypeClientPlayerInput:
			clientPlayerInput := &ClientPlayerInput{}
			if err := json.Unmarshal(message.Payload, clientPlayerInput); err != nil {
				fmt.Printf("Error unmarshalling player input: %v\n", err)
				continue
			}
			fmt.Printf("Received client player input: %+v\n", clientPlayerInput)
			gm.handlePlayerInput(message.ClientID, *clientPlayerInput)		
		
		case messages.MessageTypeClientPlayerUpdate:
			playerState := &PlayerState{}
			err := json.Unmarshal(message.Payload, playerState)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal player state: %v\n", err)
				continue
			}
			fmt.Printf("Received client player update: %+v\n", playerState)
			gm.gameState.Players[message.ClientID] = playerState
		default:
			fmt.Printf("Error: unhandled message type: %s\n", message.Type)
		}
	}

	gm.gameState.Timestamp = timestamp
}

func (gm *GameManager) handlePlayerInput(clientID uint32, input ClientPlayerInput) {
    // Log the received input for testing
    fmt.Printf("Received input from client %d: horizontal=%f, vertical=%f, jump=%v\n",
               clientID, input.Horizontal, input.Vertical, input.Jump)
	
	playerState, exists := gm.gameState.Players[clientID]
    if !exists {
        playerState = &PlayerState{/* initialize with default state */}
        gm.gameState.Players[clientID] = playerState
    }

    // Update player state based on received input
    // Example: Adjust player's position, handle jumping, etc.
    // ...

    // Implement necessary physics or game logic here
    // ...
}

func (gm *GameManager) broadcastGameState() {
	payload, err := json.Marshal(gm.gameState)
	if err != nil {
		fmt.Printf("Error: failed to marshal game state: %v\n", err)
		return
	}

	clients := gm.clientManager.GetClients()
	for _, client := range clients {
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
		// TODO: realiable vs unreliable messages
		err := servers.WriteMessageToUDP(gm.clientManager.GetUDPConn(), client.UDPAddress, message)
		if err != nil {
			fmt.Printf("Error: failed to write message to UDP connection for client %d: %v\n", client.ID, err)
		}
	}
}
