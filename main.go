package main

import (
	"fmt"
	"os"

	"github.com/ScriptMang/drumstick/internal/accts"
	"github.com/ScriptMang/drumstick/internal/backend"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func main() {

	ctx, db := backend.CreatePool()
	defer db.Close()

	var userProf accts.UserProfile
	err := pgxscan.Select(ctx, db, &userProf, `SELECT * FROM user_profile`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("welcome to Drumstick!, your user profile is %v\n", userProf)
}
