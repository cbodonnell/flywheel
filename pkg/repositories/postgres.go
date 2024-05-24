package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories/models"
	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	conn *pgx.Conn
}

// NewPostgresRepository creates a new PSQLRepository.
// It panics if it is unable to connect to the database after 2 minutes.
// The caller is responsible for calling Close() on the repository.
func NewPostgresRepository(ctx context.Context, connStr string) (Repository, error) {
	const maxRetry = 24
	const retryInterval = time.Second * 5

	var conn *pgx.Conn
	var err error

	for attempt := 1; attempt <= maxRetry; attempt++ {
		conn, err = connectDb(ctx, connStr)
		if err == nil {
			break
		}

		log.Warn("Unable to connect to the database (attempt %d): %v", attempt, err)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while connecting to database: %v", ctx.Err())
		case <-time.After(retryInterval):
			continue
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection after %d attempts: %v", maxRetry, err)
	}

	return &PostgresRepository{
		conn: conn,
	}, nil
}

func connectDb(ctx context.Context, connStr string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	var username string
	var database string
	err = conn.QueryRow(ctx, "SELECT current_user, current_database()").Scan(&username, &database)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}

	log.Info("Connected to %s as %s", database, username)

	return conn, nil
}

func (r *PostgresRepository) Close(ctx context.Context) error {
	return r.conn.Close(ctx)
}

func (r *PostgresRepository) CreateUser(ctx context.Context, userID string) (*models.User, error) {
	q := `INSERT INTO users (id) VALUES ($1) ON CONFLICT DO NOTHING;`
	_, err := r.conn.Exec(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %v", err)
	}

	return &models.User{
		ID: userID,
	}, nil
}

func (r *PostgresRepository) ListCharacters(ctx context.Context, userID string) ([]*models.Character, error) {
	q := `SELECT id, name FROM characters WHERE user_id = $1;`
	rows, err := r.conn.Query(ctx, q, userID)
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

func (r *PostgresRepository) CountCharacters(ctx context.Context, userID string) (int, error) {
	q := `SELECT COUNT(*) FROM characters WHERE user_id = $1;`
	var count int
	if err := r.conn.QueryRow(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count characters: %v", err)
	}

	return count, nil
}

func (r *PostgresRepository) GetCharacter(ctx context.Context, userID string, characterID int32) (*models.Character, error) {
	q := `SELECT id, name FROM characters WHERE id = $1 AND user_id = $2;`
	character := &models.Character{}
	if err := r.conn.QueryRow(ctx, q, characterID, userID).Scan(&character.ID, &character.Name); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &ErrNotFound{}
		}
		return nil, fmt.Errorf("failed to scan character: %v", err)
	}

	return character, nil
}

func (r *PostgresRepository) CreateCharacter(ctx context.Context, userID string, name string) (*models.Character, error) {
	q := `INSERT INTO characters (user_id, name) VALUES ($1, $2) RETURNING id;`
	var characterID int32
	if err := r.conn.QueryRow(ctx, q, userID, name).Scan(&characterID); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"characters_name_unique\"") {
			return nil, &ErrNameExists{}
		}
		return nil, fmt.Errorf("failed to insert character: %v", err)
	}

	return &models.Character{
		ID:     characterID,
		UserID: userID,
		Name:   name,
	}, nil
}

func (r *PostgresRepository) DeleteCharacter(ctx context.Context, userID string, characterID int32) error {
	q := `DELETE FROM characters WHERE id = $1 AND user_id = $2;`
	res, err := r.conn.Exec(ctx, q, characterID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete character: %v", err)
	}

	if res.RowsAffected() == 0 {
		return &ErrNotFound{}
	}

	return nil
}

func (r *PostgresRepository) SaveGameState(ctx context.Context, gameState *gametypes.GameState) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, playerState := range gameState.Players {
		q := `
		INSERT INTO players (character_id, timestamp, x, y, hitpoints) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (character_id) DO UPDATE SET timestamp = $2, x = $3, y = $4, hitpoints = $5;
		`
		_, err = tx.Exec(ctx, q, playerState.CharacterID, gameState.Timestamp, playerState.Position.X, playerState.Position.Y, playerState.Hitpoints)
		if err != nil {
			return fmt.Errorf("failed to insert player: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *PostgresRepository) SavePlayerState(ctx context.Context, timestamp int64, characterID int32, playerState *gametypes.PlayerState) error {
	q := `
	INSERT INTO players (character_id, timestamp, x, y, hitpoints) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (character_id) DO UPDATE SET timestamp = $2, x = $3, y = $4, hitpoints = $5;
	`
	_, err := r.conn.Exec(ctx, q, characterID, timestamp, playerState.Position.X, playerState.Position.Y, playerState.Hitpoints)
	if err != nil {
		return fmt.Errorf("failed to insert player: %v", err)
	}

	return nil
}

func (r *PostgresRepository) LoadPlayerState(ctx context.Context, characterID int32) (*gametypes.PlayerState, error) {
	q := `
	SELECT x, y, hitpoints FROM players WHERE character_id = $1;
	`
	var x float64
	var y float64
	var hitpoints int16
	if err := r.conn.QueryRow(ctx, q, characterID).Scan(&x, &y, &hitpoints); err != nil {
		if err == pgx.ErrNoRows {
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
