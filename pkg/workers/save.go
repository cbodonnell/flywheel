package workers

import (
	"context"
	"time"

	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/state"
)

type SaveGameStateWorker struct {
	repository          repositories.Repository
	savePlayerStateChan <-chan SavePlayerStateRequest
	stateManager        state.StateManager
	interval            time.Duration
}

type NewSaveGameStateWorkerOptions struct {
	Repository          repositories.Repository
	SavePlayerStateChan <-chan SavePlayerStateRequest
	StateManager        state.StateManager
	Interval            time.Duration
}

type SavePlayerStateRequest struct {
	Timestamp   int64
	ClientID    uint32
	PlayerState *types.PlayerState
}

// NewSaveGameStateWorker creates a new SaveGameStateWorker.
// The worker processes save requests from the game loop and
// periodically saves the game state to the repository.
func NewSaveGameStateWorker(opts NewSaveGameStateWorkerOptions) *SaveGameStateWorker {
	return &SaveGameStateWorker{
		repository:          opts.Repository,
		savePlayerStateChan: opts.SavePlayerStateChan,
		stateManager:        opts.StateManager,
		interval:            opts.Interval,
	}
}

func (w *SaveGameStateWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case saveRequest := <-w.savePlayerStateChan:
			w.savePlayerState(ctx, saveRequest)
		case t := <-ticker.C:
			gameState, err := w.stateManager.Get()
			if err != nil {
				log.Error("Failed to get current game state: %v", err)
				continue
			}
			gameState.Timestamp = t.UnixMilli()
			w.saveGameState(ctx, gameState)
		}
	}
}

func (w *SaveGameStateWorker) savePlayerState(ctx context.Context, saveRequest SavePlayerStateRequest) {
	err := w.repository.SavePlayerState(ctx, saveRequest.Timestamp, saveRequest.ClientID, saveRequest.PlayerState)
	if err != nil {
		log.Error("Failed to save player state: %v", err)
	}
}

func (w *SaveGameStateWorker) saveGameState(ctx context.Context, gameState *types.GameState) {
	err := w.repository.SaveGameState(ctx, gameState)
	if err != nil {
		log.Error("Failed to save game state: %v", err)
	}
}
