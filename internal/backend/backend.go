package backend

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func Connect() (context.Context, *pgxpool.Pool) {
	ctx := context.Background()
	uri := "postgres://username@localhost:5432/drumstick"
	os.Setenv("DATABASE_URL", uri)
	dbURL := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Unable to connect to database: %s\n", err.Error())
		os.Exit(1)
	}

	return ctx, conn
}
