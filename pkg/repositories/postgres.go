package repositories

import (
	"context"
	"fmt"
	"time"

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
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for clientID, playerState := range gameState.Players {
		q := `
		INSERT INTO players (player_id, created_at, x, y) VALUES ($1, $2, $3, $4)
		ON CONFLICT (player_id) DO UPDATE SET updated_at= $2, x = $3, y = $4;
		`
		_, err = tx.Exec(ctx, q, clientID, gameState.Timestamp, playerState.P.X, playerState.P.Y)
		if err != nil {
			return fmt.Errorf("failed to insert player: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *PostgresRepository) LoadGameState(ctx context.Context) (*gametypes.GameState, error) {
	gameState := &gametypes.GameState{
		Players: make(map[uint32]*gametypes.PlayerState),
	}

	rows, err := r.conn.Query(ctx, "SELECT player_id, x, y FROM players")
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var clientID uint32
		var x float64
		var y float64
		if err := rows.Scan(&clientID, &x, &y); err != nil {
			return nil, fmt.Errorf("failed to scan player: %v", err)
		}
		gameState.Players[clientID] = &gametypes.PlayerState{
			P: gametypes.Position{
				X: x,
				Y: y,
			},
		}
	}

	return gameState, nil
}

func (r *PostgresRepository) SavePlayerState(ctx context.Context, clientID uint32, playerState *gametypes.PlayerState) error {
	q := `
	INSERT INTO players (player_id, created_at, x, y) VALUES ($1, $2, $3, $4)
	ON CONFLICT (player_id) DO UPDATE SET updated_at= $2, x = $3, y = $4;
	`
	_, err := r.conn.Exec(ctx, q, clientID, time.Now().UnixMilli(), playerState.P.X, playerState.P.Y)
	if err != nil {
		return fmt.Errorf("failed to insert player: %v", err)
	}

	return nil
}

func (r *PostgresRepository) LoadPlayerState(ctx context.Context, clientID uint32) (*gametypes.PlayerState, error) {
	q := `
	SELECT x, y FROM players WHERE player_id = $1;
	`
	var x float64
	var y float64
	if err := r.conn.QueryRow(ctx, q, clientID).Scan(&x, &y); err != nil {
		return nil, fmt.Errorf("failed to scan player: %v", err)
	}

	return &gametypes.PlayerState{
		P: gametypes.Position{
			X: x,
			Y: y,
		},
	}, nil
}
