package workers

import (
	"context"

	"github.com/cbodonnell/flywheel/pkg/clients"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
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
			log.Error("Unknown client event type: %v", event.Type)
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
			log.Error("Failed to get player state for client %d: %v", event.ClientID, err)
		}
		log.Debug("Adding client %d with default values", event.ClientID)
		playerState = &gametypes.PlayerState{
			P: gametypes.Position{
				X: 0,
				Y: 0,
			},
		}
	}

	if err := w.connectionEventQueue.Enqueue(&gametypes.ConnectPlayerEvent{
		ClientID:    event.ClientID,
		PlayerState: playerState,
	}); err != nil {
		log.Error("Failed to enqueue connect player event: %v", err)
	}
}

func (w *ClientEventWorker) handleClientDisconnect(event clients.ClientEvent) {
	if err := w.connectionEventQueue.Enqueue(&gametypes.DisconnectPlayerEvent{
		ClientID: event.ClientID,
	}); err != nil {
		log.Error("Failed to enqueue disconnect player event: %v", err)
	}
}
