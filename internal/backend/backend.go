package backend

import (
	"context"
	"log"
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
		log.Printf("Unable to connect to database: %s\n", err.Error())
		os.Exit(1)
	}

	return ctx, conn
}
