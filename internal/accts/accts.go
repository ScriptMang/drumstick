package accts

import (
	"errors"
	"fmt"
	"scriptmang/drumstick/internal/backend"
	"strings"

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

func VetEmptyFields(acct Account) []error {

	errEmptyField := errors.New("field is empty")
	var rsltErr []error
	if acct.Fname == "" {
		rsltErr = append(rsltErr, fmt.Errorf("error:fname:%w", errEmptyField))
	}
	}

	if acct.Lname == "" {
		rsltErr = append(rsltErr, fmt.Errorf("error:lname:%w", errEmptyField))
	}
	}

	if acct.Address == "" {
		rsltErr = append(rsltErr, fmt.Errorf("error:address:%w", errEmptyField))
	}

	if acct.Username == "" {
		rsltErr = append(rsltErr, fmt.Errorf("error:username:%w", errEmptyField))
	}

	if len(acct.Password) == 0 {
		rsltErr = append(rsltErr, fmt.Errorf("error:passsword:%w", errEmptyField))
	}

	return rsltErr
}

// check user credentials for empty fields
// and append the errors to the err slice
func VetUserCreds(username, password string) []error {
	emptyUsername := errors.New("username can't be empty")
	emptyPswd := errors.New("password can't be empty")
	usernameTooShort := errors.New("username is too short")
	usernameTooLong := errors.New("username is too long")
	pswdTooShort := errors.New("password is too short")
	pswdTooLong := errors.New("password is too long")
	missingCapitalLetter := errors.New("password is missing a capital letter")
	missingNumber := errors.New("password is missing a capital letter")
	punctInUsername := errors.New("no special punctuation in the username")
	symbolsInUsername := errors.New("no special symbols in the username")

	var rsltErr []error
	switch {
	case username == "":
		rsltErr = append(rsltErr, emptyUsername)
	case len(username) < 15:
		rsltErr = append(rsltErr, usernameTooShort)
	case len(username) > 15:
		rsltErr = append(rsltErr, usernameTooLong)
	}

	// check for punct in username
	punct := " ?!;:,."
	userHasPunct := strings.ContainsAny(username, punct)
	if userHasPunct {
		rsltErr = append(rsltErr, punctInUsername)
	}
	// check for symbols in  username
	symbols := "@#$%^&*[]{}()%|\\`~"
	userHasSymbols := strings.ContainsAny(username, symbols)
	if userHasSymbols {
		rsltErr = append(rsltErr, symbolsInUsername)
	}

	switch {
	case len(password) == 0:
		rsltErr = append(rsltErr, emptyPswd)
	case len(password) < 15:
		rsltErr = append(rsltErr, pswdTooShort)
	case len(password) > 15:
		rsltErr = append(rsltErr, pswdTooLong)
	}

	// check for capital Letter in pswd
	capLetters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pswdHasCaps := strings.ContainsAny(password, capLetters)
	if !pswdHasCaps {
		rsltErr = append(rsltErr, missingCapitalLetter)
	}

	// check for number in pswd
	nums := "012345689"
	pswdHasNums := strings.ContainsAny(password, nums)
	if !pswdHasNums {
		rsltErr = append(rsltErr, missingNumber)
	}

	return rsltErr
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
