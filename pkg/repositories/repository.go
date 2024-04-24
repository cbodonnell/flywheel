package repositories

import (
	"context"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
)

type Repository interface {
	Close(ctx context.Context) error
	SaveGameState(ctx context.Context, gameState *gametypes.GameState) error
	SavePlayerState(ctx context.Context, timestamp int64, playerID string, position kinematic.Vector) error
	LoadPlayerState(ctx context.Context, playerID string) (*kinematic.Vector, error)
}
