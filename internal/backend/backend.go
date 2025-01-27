package backend

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool() (context.Context, *pgxpool.Pool) {
	ctx := context.Background()
	uri := "postgres://username@localhost:5432/drumstick"
	os.Setenv("DATABASE_URL", uri)
	db_url := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(ctx, db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n")
		os.Exit(1)
	}

	return conn, ctx
}
