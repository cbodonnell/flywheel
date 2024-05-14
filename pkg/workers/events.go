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

type ClientEventWorker struct {
	clientManager        *network.ClientManager
	repository           repositories.Repository
	connectionEventQueue queue.Queue
}

type NewClientEventWorkerOptions struct {
	ClientManager        *network.ClientManager
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
		case network.ClientEventTypeConnect:
			w.handleClientConnect(event)
		case network.ClientEventTypeDisconnect:
			w.handleClientDisconnect(event)
		default:
			log.Error("Unknown client event type: %v", event.Type)
			continue
		}
	}
}

func (w *ClientEventWorker) handleClientConnect(event network.ClientEvent) {
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

	if err := w.connectionEventQueue.Enqueue(&gametypes.ConnectPlayerEvent{
		ClientID:           event.ClientID,
		CharacterID:        character.ID,
		CharacterName:      character.Name,
		CharacterPosition:  position,
		CharacterHitpoints: hitpoints,
	}); err != nil {
		log.Error("Failed to enqueue connect player event: %v", err)
	}
}

func (w *ClientEventWorker) handleClientDisconnect(event network.ClientEvent) {
	if err := w.connectionEventQueue.Enqueue(&gametypes.DisconnectPlayerEvent{
		ClientID: event.ClientID,
	}); err != nil {
		log.Error("Failed to enqueue disconnect player event: %v", err)
	}
}
