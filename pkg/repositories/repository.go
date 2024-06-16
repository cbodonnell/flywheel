package repositories

import (
	"context"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/repositories/models"
)

type Repository interface {
	Close(ctx context.Context) error

	CreateUser(ctx context.Context, userID string) (*models.User, error)
	ListCharacters(ctx context.Context, userID string) ([]*models.Character, error)
	CountCharacters(ctx context.Context, userID string) (int, error)
	GetCharacter(ctx context.Context, userID string, characterID int32) (*models.Character, error)
	CreateCharacter(ctx context.Context, userID string, name string) (*models.Character, error)
	DeleteCharacter(ctx context.Context, userID string, characterID int32) error
	NameExists(ctx context.Context, name string) (bool, error)

	SaveGameState(ctx context.Context, gameState *gametypes.GameState) error
	SavePlayerState(ctx context.Context, timestamp int64, characterID int32, playerState *gametypes.PlayerState) error
	LoadPlayerState(ctx context.Context, characterID int32) (*gametypes.PlayerState, error)
}
