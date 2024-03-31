package state

import (
	"context"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

// StateManager provides shared access to the game state.
// Implementations must be thread-safe.
type StateManager interface {
	// Get returns a copy of the current game state.
	Get(ctx context.Context) (*gametypes.GameState, error)
	// Set sets the current game state.
	Set(ctx context.Context, gameState *gametypes.GameState) error
}
