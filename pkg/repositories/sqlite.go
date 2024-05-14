package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/repositories/models"
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

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
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

func (r *SQLiteRepository) CreateUser(ctx context.Context, userID string) (*models.User, error) {
	q := `INSERT OR IGNORE INTO users (id) VALUES (?);`
	_, err := r.db.ExecContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %v", err)
	}

	return &models.User{
		ID: userID,
	}, nil
}

func (r *SQLiteRepository) ListCharacters(ctx context.Context, userID string) ([]*models.Character, error) {
	q := `SELECT id, name FROM characters WHERE user_id = $1;`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query characters: %v", err)
	}
	defer rows.Close()

	characters := []*models.Character{}
	for rows.Next() {
		character := &models.Character{}
		if err := rows.Scan(&character.ID, &character.Name); err != nil {
			return nil, fmt.Errorf("failed to scan character: %v", err)
		}
		characters = append(characters, character)
	}

	return characters, nil
}

func (r *SQLiteRepository) CountCharacters(ctx context.Context, userID string) (int, error) {
	q := `SELECT COUNT(*) FROM characters WHERE user_id = ?;`
	var count int
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count characters: %v", err)
	}

	return count, nil
}

func (r *SQLiteRepository) GetCharacter(ctx context.Context, userID string, characterID int32) (*models.Character, error) {
	q := `SELECT id, name FROM characters WHERE id = ? AND user_id = ?;`
	character := &models.Character{}
	if err := r.db.QueryRowContext(ctx, q, characterID, userID).Scan(&character.ID, &character.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{}
		}
		return nil, fmt.Errorf("failed to scan character: %v", err)
	}

	return character, nil
}

func (r *SQLiteRepository) CreateCharacter(ctx context.Context, userID string, name string) (*models.Character, error) {
	q := `INSERT INTO characters (user_id, name) VALUES (?, ?);`
	result, err := r.db.ExecContext(ctx, q, userID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert character: %v", err)
	}

	characterID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %v", err)
	}

	return &models.Character{
		ID:   int32(characterID),
		Name: name,
	}, nil
}

func (r *SQLiteRepository) DeleteCharacter(ctx context.Context, userID string, characterID int32) error {
	q := `DELETE FROM characters WHERE id = ? AND user_id = ?;`
	result, err := r.db.ExecContext(ctx, q, characterID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete character: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return &ErrNotFound{}
	}

	return nil
}

func (r *SQLiteRepository) SaveGameState(ctx context.Context, gameState *gametypes.GameState) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, playerState := range gameState.Players {
		q := `
		INSERT OR REPLACE INTO players (character_id, timestamp, x, y, hitpoints)
		VALUES (?, ?, ?, ?, ?);
		`
		_, err = tx.ExecContext(ctx, q, playerState.CharacterID, gameState.Timestamp, playerState.Position.X, playerState.Position.Y, playerState.Hitpoints)
		if err != nil {
			return fmt.Errorf("failed to insert player: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) SavePlayerState(ctx context.Context, timestamp int64, characterID int32, playerState *gametypes.PlayerState) error {
	q := `
	INSERT OR REPLACE INTO players (character_id, timestamp, x, y, hitpoints)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, q, characterID, timestamp, playerState.Position.X, playerState.Position.Y, playerState.Hitpoints)
	if err != nil {
		return fmt.Errorf("failed to insert player: %v", err)
	}

	return nil
}

func (r *SQLiteRepository) LoadPlayerState(ctx context.Context, characterID int32) (*gametypes.PlayerState, error) {
	q := `
	SELECT x, y, hitpoints FROM players WHERE character_id = $1;
	`
	var x float64
	var y float64
	var hitpoints int16
	if err := r.db.QueryRowContext(ctx, q, characterID).Scan(&x, &y, &hitpoints); err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{}
		}
		return nil, fmt.Errorf("failed to scan player: %v", err)
	}

	return &gametypes.PlayerState{
		CharacterID: characterID,
		Position: kinematic.Vector{
			X: x,
			Y: y,
		},
		Hitpoints: hitpoints,
	}, nil
}
