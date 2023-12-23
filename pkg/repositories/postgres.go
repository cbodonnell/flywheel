package repositories

import (
	"context"
	"fmt"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	conn *pgx.Conn
}

// NewPostgresRepository creates a new PSQLRepository.
// It panics if it is unable to connect to the database.
// The caller is responsible for calling Close() on the repository.
func NewPostgresRepository(ctx context.Context, connStr string) Repository {
	return &PostgresRepository{
		conn: connectDb(ctx, connStr),
	}
}

func connectDb(ctx context.Context, connStr string) *pgx.Conn {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database: %v\n", err))
	}

	var username string
	var database string
	err = conn.QueryRow(ctx, "SELECT current_user, current_database()").Scan(&username, &database)
	if err != nil {
		panic(fmt.Sprintf("Unable to query database: %v\n", err))
	}

	fmt.Printf("Connected to %s as %s\n", database, username)

	return conn
}

func (r *PostgresRepository) Close(ctx context.Context) {
	r.conn.Close(ctx)
}

func (r *PostgresRepository) SaveGameState(ctx context.Context, gameState *gametypes.GameState) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %v\n", err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO game_states (timestamp) VALUES ($1)", gameState.Timestamp); err != nil {
		return fmt.Errorf("Failed to insert game state: %v\n", err)
	}

	for clientID, playerState := range gameState.Players {
		if _, err = tx.Exec(ctx, "INSERT INTO players (player_id) VALUES ($1) ON CONFLICT (player_id) DO NOTHING", clientID); err != nil {
			return fmt.Errorf("Failed to insert player: %v\n", err)
		}

		if _, err = tx.Exec(ctx, "INSERT INTO player_states (timestamp, player_id) VALUES ($1, $2)", gameState.Timestamp, clientID); err != nil {
			return fmt.Errorf("Failed to insert player state: %v\n", err)
		}

		if _, err = tx.Exec(ctx, "INSERT INTO player_positions (timestamp, player_id, x, y) VALUES ($1, $2, $3, $4)", gameState.Timestamp, clientID, playerState.P.X, playerState.P.Y); err != nil {
			return fmt.Errorf("Failed to insert player position: %v\n", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Failed to commit transaction: %v\n", err)
	}

	return nil
}

func (r *PostgresRepository) LoadGameState(ctx context.Context) (*gametypes.GameState, error) {
	// TODO: Implement
	return nil, nil
}
