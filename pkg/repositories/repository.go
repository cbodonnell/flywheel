package repositories

import (
	"context"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
)

type Repository interface {
	Close(ctx context.Context) error
	SaveGameState(ctx context.Context, gameState *gametypes.GameState) error
	SavePlayerState(ctx context.Context, timestamp int64, clientID uint32, playerState *gametypes.PlayerState) error
	LoadPlayerState(ctx context.Context, clientID uint32) (*gametypes.PlayerState, error)
}
