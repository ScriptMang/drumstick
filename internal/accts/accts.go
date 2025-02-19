package accts

import (
	"bytes"
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
	Email    string `json:"email" form:"email"`
	Password []byte `json:"password" form:"password"`
}

type Account struct {
	ID       int    `json:"id" form:"id"`
	Fname    string `json:"fname" form:"fname"`
	Lname    string `json:"lname" form:"lname"`
	Address  string `json:"address" form:"address"`
	Email    string `json:"email" form:"email"`
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

// function for wrapping a bunch of empty field errors
// and returning them as a slice
func fieldIsEmpty(val, fieldname string) error {
	errEmptyField := errors.New("field is empty")
	var rsltErr error
	if val == "" || len(val) == 0 {
		rsltErr = fmt.Errorf("error:%s:%w", fieldname, errEmptyField)
	}
	return rsltErr
}

// checks an account field for any numbers and returns a slice of errors
func fieldHasNumbers(val, fieldname string) error {
	errHasNums := errors.New("field can't contain any numbers")
	var rsltErr error
	if strings.ContainsAny(val, "0123456789") {
		rsltErr = fmt.Errorf("error:%s:%w", fieldname, errHasNums)
	}
	return rsltErr
}

// checks an account field for any symbols and returns a slice of errors
func fieldHasSymbols(val, fieldname string) error {
	var rsltErr error
	errHasSymbols := errors.New("field can't contain any symbols")
	symbolsFilter := "!@$_^%&*();/-+=\"'`~[]{}<|>"

	if strings.ContainsAny(val, symbolsFilter) {
		rsltErr = fmt.Errorf("error:%s:%w", fieldname, errHasSymbols)
	}
	return rsltErr
}

// performs mult. error checks to validate an
// email address and returns the final error
func vetEmailAddress(email string) error {
	var tmpErrs []error

	reqSymbols := "@"
	endingAddrs := []string{".com", ".org", ".net"}
	reqNums := "0123456789"
	errReqNums := errors.New("email requires numbers.")
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// the email must have contain a numer
	if !strings.ContainsAny(email, reqNums) {
		tmpErrs = append(tmpErrs, fmt.Errorf("error:email:%w", errReqNums))
	}

	// the  email must have an @ symbol
	if !strings.ContainsAny(email, reqSymbols) {
		tmpErrs = append(tmpErrs, fmt.Errorf("error:email:%w", errReqSymbol))
	}

	validEmailOrg := false
	for _, endingAddr := range endingAddrs {
		if strings.Contains(email, endingAddr) {
			validEmailOrg = true
		}
	}

	if !validEmailOrg {
		tmpErrs = append(tmpErrs, fmt.Errorf("error:email:%w", errReqEndingAddr))
	}

	return errors.Join(tmpErrs...)
}

func VetAllFields(acct Account) []error {
	var tmpErrs []error
	var rsltErr []error

	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Lname, "lname"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Address, "adress"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Email, "email"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(string(acct.Password), "password"))

	tmpErrs = append(tmpErrs, fieldHasNumbers(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasNumbers(acct.Lname, "lname"))

	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Lname, "lname"))
	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Address, "adress"))
	tmpErrs = append(tmpErrs, vetEmailAddress(acct.Email))

	symbolsFilter := "!@$_^%&*();/-+=\"'`~[]{}<|>"
	errHasSymbols := errors.New("field can't contain any symbols")
	if bytes.ContainsAny(acct.Password, symbolsFilter) {
		tmpErrs = append(tmpErrs, fmt.Errorf("error:password:%w", errHasSymbols))
	}

	for _, err := range tmpErrs {
		if err != nil {
			rsltErr = append(rsltErr, err)
		}
	}

	return rsltErr
}

// check user credentials for empty fields
// and append the errors to the err slice
func VetUserCreds(email, password string) []error {
	emptyEmail := errors.New("email can't be empty")
	emptyPswd := errors.New("password can't be empty")
	emailTooShort := errors.New("email is too short")
	emailTooLong := errors.New("email is too long")
	pswdTooShort := errors.New("password is too short")
	pswdTooLong := errors.New("password is too long")
	missingCapitalLetter := errors.New("password is missing a capital letter")
	missingNumber := errors.New("password is missing a capital letter")
	punctInEmail := errors.New("no special punctuation in the email")
	symbolsInEmail := errors.New("no special symbols in the email")

	var rsltErr []error
	switch {
	case email == "":
		rsltErr = append(rsltErr, emptyEmail)
	case len(email) < 15:
		rsltErr = append(rsltErr, emailTooShort)
	case len(email) > 15:
		rsltErr = append(rsltErr, emailTooLong)
	}

	// check for punct in email
	punct := " ?!;:,"
	emailHasPunct := strings.ContainsAny(email, punct)
	if emailHasPunct {
		rsltErr = append(rsltErr, punctInEmail)
	}
	// check for symbols in  email
	symbols := "#$%^&*[]{}()%|\\`~"
	userHasSymbols := strings.ContainsAny(email, symbols)
	if userHasSymbols {
		rsltErr = append(rsltErr, symbolsInEmail)
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
		`INSERT INTO user_account(email, password) VALUES($1, $2) RETURNING id`,
		acct.Email, acct.Password,
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
func UserIDByEmail(email string) (int, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var possibleUserID []*int
	err := pgxscan.Select(ctx, db, &possibleUserID, `SELECT id FROM user_account WHERE email = $1`, email)
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
	userID, err = UserIDByEmail(acct.Email)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	err = addUserProfile(acct)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	return fmt.Sprintf("user %d has been registered", userID), nil
}
