package workers

import (
	"context"
	"time"

	"github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
)

type SaveGameStateWorker struct {
	repository          repositories.Repository
	savePlayerStateChan <-chan SavePlayerStateRequest
	gameState           *types.GameState
	interval            time.Duration
}

type NewSaveGameStateWorkerOptions struct {
	Repository          repositories.Repository
	SavePlayerStateChan <-chan SavePlayerStateRequest
	GameState           *types.GameState
	Interval            time.Duration
}

type SavePlayerStateRequest struct {
	Timestamp   int64
	CharacterID int32
	PlayerState *types.PlayerState
}

// NewSaveGameStateWorker creates a new SaveGameStateWorker.
// The worker processes save requests from the game loop and
// periodically saves the game state to the repository.
func NewSaveGameStateWorker(opts NewSaveGameStateWorkerOptions) *SaveGameStateWorker {
	return &SaveGameStateWorker{
		repository:          opts.Repository,
		savePlayerStateChan: opts.SavePlayerStateChan,
		gameState:           opts.GameState,
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
		case <-ticker.C:
			gameState := w.gameState.Copy()
			w.saveGameState(ctx, gameState)
		}
	}
}

func (w *SaveGameStateWorker) savePlayerState(ctx context.Context, saveRequest SavePlayerStateRequest) {
	err := w.repository.SavePlayerState(ctx, saveRequest.Timestamp, saveRequest.CharacterID, saveRequest.PlayerState)
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
