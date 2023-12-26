package state

import (
	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

// StateManager provides shared access to the game state.
type StateManager interface {
	// Get returns a copy of the current game state.
	Get() (*gametypes.GameState, error)
	// Set sets the current game state.
	Set(gameState *gametypes.GameState) error
}
