package workers

import (
	"context"
	"fmt"

	"github.com/cbodonnell/flywheel/pkg/clients"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
)

type ClientEventWorker struct {
	clientManager        *clients.ClientManager
	repository           repositories.Repository
	connectionEventQueue queue.Queue
}

type NewClientEventWorkerOptions struct {
	ClientManager        *clients.ClientManager
	Repository           repositories.Repository
	ConnectionEventQueue queue.Queue
}

// NewClientEventWorker creates a new ClientEventWorker.
// The worker processes client events like connect and disconnect
// and writes connection events to a queue for the game loop to process.
func NewClientEventWorker(opts NewClientEventWorkerOptions) *ClientEventWorker {
	return &ClientEventWorker{
		clientManager:        opts.ClientManager,
		repository:           opts.Repository,
		connectionEventQueue: opts.ConnectionEventQueue,
	}
}

func (w *ClientEventWorker) Start() {
	for event := range w.clientManager.GetClientEventChan() {
		switch event.Type {
		case clients.ClientEventTypeConnect:
			w.handleClientConnect(event)
		case clients.ClientEventTypeDisconnect:
			w.handleClientDisconnect(event)
		default:
			fmt.Printf("Error: unknown client event type: %v\n", event.Type)
			continue
		}
	}
}

func (w *ClientEventWorker) handleClientConnect(event clients.ClientEvent) {
	var playerState *gametypes.PlayerState
	if lastKnownState, err := w.repository.LoadPlayerState(context.Background(), event.ClientID); err == nil {
		playerState = lastKnownState
	} else {
		if !repositories.IsNotFound(err) {
			fmt.Printf("Error: failed to get player state for client %d: %v\n", event.ClientID, err)
		}
		fmt.Printf("Adding client %d with default values\n", event.ClientID)
		playerState = &gametypes.PlayerState{
			P: gametypes.Position{
				X: 0,
				Y: 0,
			},
		}
	}

	w.connectionEventQueue.Enqueue(&gametypes.ConnectPlayerEvent{
		ClientID:    event.ClientID,
		PlayerState: playerState,
	})
}

func (w *ClientEventWorker) handleClientDisconnect(event clients.ClientEvent) {
	w.connectionEventQueue.Enqueue(&gametypes.DisconnectPlayerEvent{
		ClientID: event.ClientID,
	})
}
