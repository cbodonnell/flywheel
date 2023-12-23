package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	conn *pgx.Conn
}

// NewPostgresRepository creates a new PSQLRepository.
// It panics if it is unable to connect to the database.
// The caller is responsible for calling Close() on the repository.
func NewPostgresRepository(connStr string) Repository {
	return &PostgresRepository{
		conn: connectDb(connStr),
	}
}

func connectDb(connStr string) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database: %v\n", err))
	}

	var username string
	var database string
	err = conn.QueryRow(context.Background(), "SELECT current_user, current_database()").Scan(&username, &database)
	if err != nil {
		panic(fmt.Sprintf("Unable to query database: %v\n", err))
	}

	fmt.Printf("Connected to %s as %s\n", database, username)

	return conn
}

func (r *PostgresRepository) Close() {
	r.conn.Close(context.Background())
}
