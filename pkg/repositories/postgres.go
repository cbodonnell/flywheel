package repositories

import (
	"context"
	"fmt"
	"time"

	gametypes "github.com/cbodonnell/flywheel/pkg/game/types"
	"github.com/cbodonnell/flywheel/pkg/kinematic"
	"github.com/cbodonnell/flywheel/pkg/log"
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

func (r *PostgresRepository) SaveGameState(ctx context.Context, gameState *gametypes.GameState) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, playerState := range gameState.Players {
		q := `
		INSERT INTO players (player_id, timestamp, x, y) VALUES ($1, $2, $3, $4)
		ON CONFLICT (player_id) DO UPDATE SET timestamp = $2, x = $3, y = $4;
		`
		_, err = tx.Exec(ctx, q, playerState.PlayerID, gameState.Timestamp, playerState.Position.X, playerState.Position.Y)
		if err != nil {
			return fmt.Errorf("failed to insert player: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *PostgresRepository) SavePlayerState(ctx context.Context, timestamp int64, playerID string, position kinematic.Vector) error {
	q := `
	INSERT INTO players (player_id, timestamp, x, y) VALUES ($1, $2, $3, $4)
	ON CONFLICT (player_id) DO UPDATE SET timestamp = $2, x = $3, y = $4;
	`
	_, err := r.conn.Exec(ctx, q, playerID, timestamp, position.X, position.Y)
	if err != nil {
		return fmt.Errorf("failed to insert player: %v", err)
	}

	return nil
}

func (r *PostgresRepository) LoadPlayerState(ctx context.Context, playerID string) (*kinematic.Vector, error) {
	q := `
	SELECT x, y FROM players WHERE player_id = $1;
	`
	var x float64
	var y float64
	if err := r.conn.QueryRow(ctx, q, playerID).Scan(&x, &y); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &ErrNotFound{}
		}
		return nil, fmt.Errorf("failed to scan player: %v", err)
	}

	return &kinematic.Vector{
		X: x,
		Y: y,
	}, nil
}
