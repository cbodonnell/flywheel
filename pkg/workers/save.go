package workers

import (
	"context"
	"fmt"

	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
)

type SaveGameStateWorker struct {
	repository    repositories.Repository
	saveStateChan <-chan SaveStateRequest
}

type NewSaveGameStateWorkerOptions struct {
	Repository    repositories.Repository
	SaveStateChan <-chan SaveStateRequest
}

type SaveStateRequest struct {
	Timestamp int64
	Type      SaveStateRequestType
	State     interface{}
}

type SaveStateRequestType int

const (
	SaveStateRequestTypePlayer SaveStateRequestType = iota
	SaveStateRequestTypeGame
)

// NewSaveGameStateWorker creates a new SaveGameStateWorker.
// The worker processes save requests from the game loop and
// periodically saves the game state to the repository.
func NewSaveGameStateWorker(opts NewSaveGameStateWorkerOptions) *SaveGameStateWorker {
	return &SaveGameStateWorker{
		repository:    opts.Repository,
		saveStateChan: opts.SaveStateChan,
	}
}

func (w *SaveGameStateWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case saveRequest := <-w.saveStateChan:
			switch saveRequest.Type {
			case SaveStateRequestTypePlayer:
				if err := w.savePlayerState(ctx, saveRequest); err != nil {
					log.Error("Failed to save player state: %v", err)
				}
			case SaveStateRequestTypeGame:
				if err := w.saveGameState(ctx, saveRequest); err != nil {
					log.Error("Failed to save game state: %v", err)
				}
			}
		}
	}
}

func (w *SaveGameStateWorker) savePlayerState(ctx context.Context, saveRequest SaveStateRequest) error {
	playerState, ok := saveRequest.State.(*types.PlayerState)
	if !ok {
		return fmt.Errorf("failed to cast player state")
	}

	err := w.repository.SavePlayerState(ctx, saveRequest.Timestamp, playerState.CharacterID, playerState)
	if err != nil {
		return fmt.Errorf("failed to save player state: %v", err)
	}

	return nil
}

func (w *SaveGameStateWorker) saveGameState(ctx context.Context, saveRequest SaveStateRequest) error {
	gameState, ok := saveRequest.State.(*types.GameState)
	if !ok {
		return fmt.Errorf("failed to cast game state")
	}

	err := w.repository.SaveGameState(ctx, gameState)
	if err != nil {
		return fmt.Errorf("failed to save game state: %v", err)
	}

	return nil
}
