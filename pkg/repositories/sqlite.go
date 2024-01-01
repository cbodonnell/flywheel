package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(ctx context.Context, path string, migrations string) (Repository, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	dir, err := os.ReadDir(migrations)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %v", err)
	}

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		migrationPath := filepath.Join(migrations, entry.Name())
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %v", migrationPath, err)
		}

		if _, err := db.ExecContext(ctx, string(migration)); err != nil {
			return nil, fmt.Errorf("failed to execute migration %s: %v", migrationPath, err)
		}
	}

	return &SQLiteRepository{
		db: db,
	}, nil
}

func (r *SQLiteRepository) Close(ctx context.Context) error {
	return r.db.Close()
}

func (r *SQLiteRepository) SaveGameState(ctx context.Context, gameState *gametypes.GameState) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for clientID, playerState := range gameState.Players {
		q := `
		INSERT OR REPLACE INTO players (player_id, timestamp, x, y)
		VALUES (?, ?, ?, ?);
		`
		_, err = tx.ExecContext(ctx, q, clientID, gameState.Timestamp, playerState.P.X, playerState.P.Y)
		if err != nil {
			return fmt.Errorf("failed to insert player: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) SavePlayerState(ctx context.Context, timestamp int64, clientID uint32, playerState *gametypes.PlayerState) error {
	q := `
	INSERT OR REPLACE INTO players (player_id, timestamp, x, y)
	VALUES (?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, q, clientID, timestamp, playerState.P.X, playerState.P.Y)
	if err != nil {
		return fmt.Errorf("failed to insert player: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) LoadPlayerState(ctx context.Context, clientID uint32) (*gametypes.PlayerState, error) {
	q := `
	SELECT x, y FROM players WHERE player_id = $1;
	`
	var x float64
	var y float64
	if err := r.db.QueryRowContext(ctx, q, clientID).Scan(&x, &y); err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{}
		}
		return nil, fmt.Errorf("failed to scan player: %v", err)
	}

	return &gametypes.PlayerState{
		P: gametypes.Position{
			X: x,
			Y: y,
		},
	}, nil
}
