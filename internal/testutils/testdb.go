package testutils

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("env var TEST_DATABASE_URL  is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func ResetAndTestTx(t *testing.T, db *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()

	_, resetErr := db.Exec(ctx, "TRUNCATE user_account RESTART IDENTITY CASCADE")
	if resetErr != nil {
		t.Fatal(resetErr)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	fn(ctx, tx)
}
