package game

import (
	"fmt"
	"time"

	"github.com/cbodonnell/flywheel/pkg/clients"
	"github.com/cbodonnell/flywheel/pkg/messages"
	"github.com/cbodonnell/flywheel/pkg/queue" // Import the path to your queue package
)

// GameState represents the generic game state (to be implemented based on your needs)
type GameState struct {
	// Add fields based on your game's state
}

// GameManager represents the game manager.
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
		gameState:     &GameState{}, // Initialize your game state here
		loopInterval:  loopInterval,
		stopChannel:   make(chan struct{}),
	}
}

// StartGameLoop starts the game loop.
func (gm *GameManager) StartGameLoop() {
	ticker := time.NewTicker(gm.loopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Game loop logic here
			gm.processMessages()
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

func (gm *GameManager) processMessages() {
	pendingMessages := gm.messageQueue.ReadAllMessages()
	for _, item := range pendingMessages {
		message, ok := item.(messages.Message)
		if !ok {
			fmt.Println("Error: failed to cast message to messages.Message")
			continue
		}
		fmt.Printf("Received message: %+v\n", message)

		// TODO: handle message based on its type
		// this will update the game state
	}
}

func (gm *GameManager) broadcastGameState() {
	clients := gm.clientManager.GetClients()
	for _, client := range clients {
		message := messages.Message{
			ClientID: client.ID,
			Type:     messages.MessageTypeClientUpdate,
			Payload:  gm.gameState,
		}

		if client.UDPConn == nil {
			fmt.Printf("Error: client %d does not have a UDP connection\n", client.ID)
			continue
		}
		// TODO: some messages will be over UDP, some over TCP
		err := messages.WriteMessageToUDP(client.UDPConn, message)
		if err != nil {
			fmt.Printf("Error: failed to write message to UDP connection for client %d: %v\n", client.ID, err)
		}
	}
}
