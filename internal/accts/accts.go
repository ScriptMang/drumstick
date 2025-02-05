package accts

import (
	"errors"
	"fmt"
	"scriptmang/drumstick/internal/backend"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type UserProfile struct {
	ID      int    `json:"id" form:"id"`
	UserID  int    `json:"user_id" form:"user_id"`
	Fname   string `json:"fname" form:"fname"`
	Lname   string `json:"lname" form:"lname"`
	Address string `json:"address" form:"address"`
}

type UserAccount struct {
	ID       int    `json:"id" form:"id"`
	Username string `json:"username" form:"username"`
	Password []byte `json:"password" form:"password"`
}

type Account struct {
	ID       int    `json:"id" form:"id"`
	Fname    string `json:"fname" form:"fname"`
	Lname    string `json:"lname" form:"lname"`
	Address  string `json:"address" form:"address"`
	Username string `json:"username" form:"username"`
	Password []byte `json:"password" form:"password"`
}

type Posts struct {
	ID            int    `json:"id" form:"id"`
	UserID        int    `json:"user_id" form:"user_id"`
	Content       string `json:"content" form:"content"`
	NumbComments  int    `json:"number_comments" form:"number_comments"`
	NumbReposts   int    `json:"number_reposts" form:"number_reposts"`
	NumbLikes     int    `json:"number_likes" form:"number_likes"`
	NumbViews     int    `json:"number_views" form:"number_views"`
	NumbBookmarks int    `json:"number_bookmarks" form:"number_bookmarks"`
}

// encrypts a byte slice secret using bcrypt
func encryptPassword(s []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(s, 14)
	return hash, err
}

// add user account to the database
func addUserAcct(acct *Account) error {
	ctx, db := backend.Connect()
	defer db.Close()

	var err error
	acct.Password, err = encryptPassword(acct.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Code)
			fmt.Println(pgErr.Message)
		}

		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			fmt.Println(err)
		}

		if errors.Is(err, bcrypt.ErrHashTooShort) {
			fmt.Println(err)
		}

		return fmt.Errorf("error: %s, code: %s", pgErr.Message, pgErr.Code)
	}

	var tempID int
	err = db.QueryRow(ctx,
		`INSERT INTO user_account(username, password) VALUES($1, $2) RETURNING id`,
		acct.Username, acct.Password,
	).Scan(&tempID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Code)
			fmt.Println(pgErr.Message)
		}
		return fmt.Errorf("error: %s, code: %s", pgErr.Message, pgErr.Code)
	}

	acct.ID = tempID
	return nil
}

// get the user's id from their username
func UserIDByUsername(username string) (int, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var possibleUserID []*int
	err := pgxscan.Select(ctx, db, &possibleUserID, `SELECT id FROM user_account WHERE username = $1`, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("error: resource not found: id does not exist")
		}
		return 0, fmt.Errorf("error: %v", err)
	}
	userID := *possibleUserID[0]
	return userID, nil
}

// add user account to the database
func addUserProfile(acct Account) error {
	ctx, db := backend.Connect()
	defer db.Close()

	_, err := db.Exec(ctx,
		`INSERT INTO user_profile (user_id, fname, lname, address) VALUES($1,$2,$3,$4)`,
		acct.ID, acct.Fname, acct.Lname, acct.Address,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			fmt.Println(pgErr.Code)
			fmt.Println(pgErr.Message)
		}
		return fmt.Errorf("error: %s, code: %s", pgErr.Message, pgErr.Code)
	}
	return nil
}

// add a user account and profile to the database
func CreateAcct(acct Account) (string, error) {
	err := addUserAcct(&acct)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	var userID int
	userID, err = UserIDByUsername(acct.Username)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	err = addUserProfile(acct)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	return fmt.Sprintf("user %d has been registered", userID), nil
}
