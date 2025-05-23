package backend

import (
	"testing"
)

func Test_connection(t *testing.T) {
	ctx, db := Connect()
	defer db.Close()

	failedDBConn := db.Ping(ctx)
	if failedDBConn != nil {
		t.Fatalf("server error: %s\n", failedDBConn.Error())
	}
}
