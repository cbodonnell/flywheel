package workers

import (
	"context"

	gameconstants "github.com/cbodonnell/flywheel/pkg/game/constants"
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/network"
	"github.com/cbodonnell/flywheel/pkg/queue"
	"github.com/cbodonnell/flywheel/pkg/repositories"
)

type ConnectionEventWorker struct {
	connectionEventChan <-chan network.ConnectionEvent
	repository          repositories.Repository
	serverEventQueue    queue.Queue
}

type NewConnectionEventWorkerOptions struct {
	ConnectionEventChan <-chan network.ConnectionEvent
	Repository          repositories.Repository
	ServerEventQueue    queue.Queue
}

// NewConnectionEventWorker creates a new ConnectionEventWorker.
// The worker processes client events like connect and disconnect
// and writes server events to a queue for the game loop to process.
func NewConnectionEventWorker(opts NewConnectionEventWorkerOptions) *ConnectionEventWorker {
	return &ConnectionEventWorker{
		connectionEventChan: opts.ConnectionEventChan,
		repository:          opts.Repository,
		serverEventQueue:    opts.ServerEventQueue,
	}
}

func (w *ConnectionEventWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-w.connectionEventChan:
			switch event.Type {
			case network.ConnectionEventTypeConnect:
				w.handleClientConnect(event)
			case network.ConnectionEventTypeDisconnect:
				w.handleClientDisconnect(event)
			default:
				log.Error("Unknown client event type: %v", event.Type)
			}
		}
	}
}

func (w *ConnectionEventWorker) handleClientConnect(event network.ConnectionEvent) {
	data, ok := event.Data.(network.ClientConnectData)
	if !ok {
		log.Error("Failed to cast client connect data")
		return
	}

	character, err := w.repository.GetCharacter(context.Background(), data.UserID, data.CharacterID)
	if err != nil {
		log.Error("Failed to get character %d for user %s: %v", data.CharacterID, data.UserID, err)
		return
	}

	var position kinematic.Vector
	var hitpoints int16
	if lastKnownState, err := w.repository.LoadPlayerState(context.Background(), character.ID); err == nil {
		position = lastKnownState.Position
		hitpoints = lastKnownState.Hitpoints
	} else {
		if !repositories.IsNotFound(err) {
			log.Error("Failed to get player state for character %d: %v", character.ID, err)
		}
		log.Debug("Adding character %d with default values", character.ID)
		position = kinematic.Vector{
			X: gameconstants.PlayerStartingX,
			Y: gameconstants.PlayerStartingY,
		}
		hitpoints = gameconstants.PlayerHitpoints
	}

	if err := w.serverEventQueue.Enqueue(&gametypes.ConnectPlayerEvent{
		ClientID:           event.ClientID,
		CharacterID:        character.ID,
		CharacterName:      character.Name,
		CharacterPosition:  position,
		CharacterHitpoints: hitpoints,
	}); err != nil {
		log.Error("Failed to enqueue connect player event: %v", err)
	}
}

func (w *ConnectionEventWorker) handleClientDisconnect(event network.ConnectionEvent) {
	if err := w.serverEventQueue.Enqueue(&gametypes.DisconnectPlayerEvent{
		ClientID: event.ClientID,
	}); err != nil {
		log.Error("Failed to enqueue disconnect player event: %v", err)
	}
}
