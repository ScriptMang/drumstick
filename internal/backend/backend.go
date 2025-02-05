package backend

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() (context.Context, *pgxpool.Pool) {
	ctx := context.Background()
	uri := "postgres://username@localhost:5432/drumstick"
	os.Setenv("DATABASE_URL", uri)
	dbURL := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return ctx, conn
}
