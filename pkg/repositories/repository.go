package repositories

import (
	"context"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

type Repository interface {
	Close(ctx context.Context)
	SaveGameState(ctx context.Context, gameState *gametypes.GameState) error
	LoadGameState(ctx context.Context) (*gametypes.GameState, error)
}
