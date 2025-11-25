package accts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
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
	ID        int      `json:"id" form:"id"`
	Username  string   `json:"username" form:"username"`
	Email     string   `json:"email" form:"email"`
	Password  []byte   `json:"password" form:"password"`
	Followers []string `json:"followers" form:"followers"`
}

type Account struct {
	ID       int    `json:"id,omitempty" form:"id"`
	Fname    string `json:"fname" form:"fname"`
	Lname    string `json:"lname" form:"lname"`
	Username string `json:"username" form:"username"`
	Address  string `json:"address" form:"address"`
	Email    string `json:"email" form:"email"`
	Password []byte `json:"password" form:"password"`
}

// all the errors
var InvalidUserCreds error = errors.New("incorrect email or password.")

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

func fieldHasPunct(val, fieldname string) error {
	errHasPunct := errors.New("field has punctuation")
	var rsltErr error
	if strings.ContainsAny(val, " ?.!:,;") {
		rsltErr = fmt.Errorf("error:%s:%w", fieldname, errHasPunct)
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
func vetEmailAddress(email string) []error {
	var tmpErrs []error

	reqSymbols := "@"
	endingAddrs := []string{".com", ".org", ".net"}
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// if the email is empty, don't return an error
	// its already handled by another func
	if email == "" {
		return []error{}
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

	return tmpErrs
}

// verifies if the user's password meet standards for account creation
// if it doesn't a slice of errors are returned
func vetPassword(password string) []error {
	emptyPswd := errors.New("field can't be empty")
	pswdTooShort := errors.New("field too short")
	pswdTooLong := errors.New("field too long")
	pswdHasNoCapLetters := errors.New("field is missing at least 1 capital letter")
	pswdHasNoDigits := errors.New("field missing at least 1 digit")
	pswdHasSymbols := errors.New("field can't contain any symbols")

	var rsltErr []error
	switch {
	case len(password) == 0:
		rsltErr = append(rsltErr, fmt.Errorf("error:password:%w", emptyPswd))
		return rsltErr
	case len(password) < 8:
		rsltErr = append(rsltErr, fmt.Errorf("error:password:%w", pswdTooShort))
	case len(password) > 32:
		rsltErr = append(rsltErr, fmt.Errorf("error:password:%w", pswdTooLong))
	}

	// check for symbols in the pswd
	symbolsFilter := "!@$_^%&*();/-+=\"'`~[]{}<|>"
	if bytes.ContainsAny([]byte(password), symbolsFilter) {
		rsltErr = append(rsltErr, fmt.Errorf("error:password:%w", pswdHasSymbols))
	}

	// check for capital Letter in pswd
	capLetters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pswdHasCaps := strings.ContainsAny(password, capLetters)
	if !pswdHasCaps {
		rsltErr = append(rsltErr, fmt.Errorf("error:password:%w", pswdHasNoCapLetters))
	}

	// check for number in pswd
	nums := "012345689"
	pswdHasNums := strings.ContainsAny(password, nums)
	if !pswdHasNums {
		rsltErr = append(rsltErr, fmt.Errorf("error:password:%w", pswdHasNoDigits))
	}

	return rsltErr
}

func VetAllFields(acct Account) []error {
	var tmpErrs []error
	var rsltErr []error

	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Lname, "lname"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Address, "address"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Username, "username"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Email, "email"))

	tmpErrs = append(tmpErrs, fieldHasPunct(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasPunct(acct.Lname, "lname"))

	tmpErrs = append(tmpErrs, fieldHasNumbers(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasNumbers(acct.Lname, "lname"))

	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Lname, "lname"))
	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Address, "address"))
	tmpErrs = append(tmpErrs, vetEmailAddress(acct.Email)...)
	tmpErrs = append(tmpErrs, vetPassword(string(acct.Password))...)

	for _, err := range tmpErrs {
		if err != nil {
			rsltErr = append(rsltErr, err)
		}
	}

	return rsltErr
}

// returns a slice of all the emails in the database
func readEmails(ctx context.Context, db backend.Querier) []string {
	var users []*UserAccount
	err := pgxscan.Select(ctx, db, &users, `SELECT * FROM user_account`)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Println(err)
		return []string{""}
	}

	if err != nil {
		log.Println(err)
		return []string{""}
	}

	var emails []string
	for _, user := range users {
		emails = append(emails, user.Email)
	}

	return emails
}

// compare the provided user email ,and password
// hash to those that exist in the databse
func CompareUserCreds(ctx context.Context, db backend.Querier, email string, pswd string) error {
	lstEmails := readEmails(ctx, db)
	emailIsReal := false
	for _, target := range lstEmails {
		if email == target {
			emailIsReal = true
			break
		}
	}

	if !emailIsReal {
		return InvalidUserCreds
	}

	// need pswd hash from database
	// compare to one in database
	var users []*UserAccount
	err := pgxscan.Select(ctx, db, &users, `Select * FROM user_account`)

	if err != nil {
		errFailedDBEntry := errors.New("database error: couldn’t connect to drumstick")
		log.Println(errFailedDBEntry)
		return errFailedDBEntry
	}

	if len(users) == 0 {
		log.Println(InvalidUserCreds)
		return InvalidUserCreds
	}

	hashIsReal := false
	for _, user := range users {
		if bcrypt.CompareHashAndPassword(user.Password, []byte(pswd)) == nil {
			hashIsReal = true
			break
		}
	}

	if !hashIsReal {
		return InvalidUserCreds
	}

	return nil
}

// add user account to the database
func addUserAcct(ctx context.Context, db backend.Querier, acct *Account) error {
	var err error
	acct.Password, err = encryptPassword(acct.Password)
	if err != nil {

		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			// fmt.Println(err)
			return err
		}

		if errors.Is(err, bcrypt.ErrHashTooShort) {
			// fmt.Println(err)
			return err
		}

		return fmt.Errorf("%w", err)
	}

	var tempID int
	err = db.QueryRow(ctx,
		`INSERT INTO user_account(username, email, password) VALUES($1, $2, $3) RETURNING id`,
		acct.Username, acct.Email, acct.Password,
	).Scan(&tempID)

	if err != nil {
		var connErr *pgconn.ConnectError
		if errors.As(err, &connErr) {
			return errors.New("db_auth:bad_creds:failed to connect to database")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return errors.New(pgErr.Message)
		}
	}

	acct.ID = tempID
	return nil
}

// get the user's id from their username
func UserIDByEmail(ctx context.Context, db backend.Querier, email string) (int, error) {
	var user []*UserAccount
	err := pgxscan.Select(ctx, db, &user, `SELECT id FROM user_account WHERE email = $1`, email)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errors.New("error: resource not found: id does not exist")
	}

	if err != nil {
		return 0, fmt.Errorf("error: %w", err)
	}

	if len(user) == 0 {
		return 0, errors.New("error: resource not found: id does not exist")
	}

	userID := user[0].ID
	return userID, nil
}

// add user account to the database
func addUserProfile(ctx context.Context, db backend.Querier, acct Account) error {
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
func CreateAcct(ctx context.Context, db backend.Querier, acct Account) (string, error) {
	err := addUserAcct(ctx, db, &acct)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	var userID int
	userID, err = UserIDByEmail(ctx, db, acct.Email)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	err = addUserProfile(ctx, db, acct)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}

	return fmt.Sprintf("user %d has been registered", userID), nil
}
