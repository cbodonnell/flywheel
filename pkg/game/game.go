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
type ClientPlayerUpdate struct {
	// Timestamp is the client time at which position is recorded
	Timestamp   int64       `json:"timestamp"`
	PlayerState PlayerState `json:"playerState"`
}

type PlayerState struct {
	P Position `json:"p"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
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
		fmt.Printf("Received message: %+v\n", message.Type)

		switch message.Type {
		case messages.MessageTypeClientPlayerUpdate:
			clientPlayerUpdate := &ClientPlayerUpdate{}
			err := json.Unmarshal(message.Payload, &clientPlayerUpdate)
			if err != nil {
				fmt.Printf("Error: failed to unmarshal player state: %v\n", err)
				continue
			}
			gm.gameState.Players[message.ClientID] = &clientPlayerUpdate.PlayerState
		default:
			fmt.Printf("Error: unhandled message type: %s\n", message.Type)
		}
	}

	gm.gameState.Timestamp = timestamp
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
		} else {
			fmt.Printf("Sent message: %s\n", message.Type)
		}
	}
}
