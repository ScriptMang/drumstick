package accts

import (
	"bytes"
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
	Email     string   `json:"email" form:"email"`
	Password  []byte   `json:"password" form:"password"`
	Followers []string `json:"followers" form:"followers"`
}

type Account struct {
	ID       int    `json:"id,omitempty" form:"id"`
	Fname    string `json:"fname" form:"fname"`
	Lname    string `json:"lname" form:"lname"`
	Address  string `json:"address" form:"address"`
	Email    string `json:"email" form:"email"`
	Password []byte `json:"password" form:"password"`
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
	reqNums := "0123456789"
	errReqNums := errors.New("email requires numbers.")
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// the email must have contain a number
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

	return tmpErrs
}

// verifies if the user's password meet standards for account creation
// if it doesn't a slice of errors are returned
func vetPassword(password string) []error {
	emptyPswd := errors.New("password can't be empty")
	pswdTooShort := errors.New("password is too short")
	pswdTooLong := errors.New("password is too long")
	pswdHasNoCapLetters := errors.New("password is missing a capital letter")
	pswdHasNoDigits := errors.New("password is missing digits")
	pswdHasSymbols := errors.New("password can't contain any symbols")

	var rsltErr []error
	switch {
	case len(password) == 0:
		rsltErr = append(rsltErr, emptyPswd)
	case len(password) < 15:
		rsltErr = append(rsltErr, pswdTooShort)
	case len(password) > 15:
		rsltErr = append(rsltErr, pswdTooLong)
	}

	// check for symbols in the pswd
	symbolsFilter := "!@$_^%&*();/-+=\"'`~[]{}<|>"
	if bytes.ContainsAny([]byte(password), symbolsFilter) {
		rsltErr = append(rsltErr, pswdHasSymbols)
	}

	// check for capital Letter in pswd
	capLetters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pswdHasCaps := strings.ContainsAny(password, capLetters)
	if !pswdHasCaps {
		rsltErr = append(rsltErr, pswdHasNoCapLetters)
	}

	// check for number in pswd
	nums := "012345689"
	pswdHasNums := strings.ContainsAny(password, nums)
	if !pswdHasNums {
		rsltErr = append(rsltErr, pswdHasNoDigits)
	}

	return rsltErr
}

func VetAllFields(acct Account) []error {
	var tmpErrs []error
	var rsltErr []error

	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Lname, "lname"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Address, "address"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(acct.Email, "email"))
	tmpErrs = append(tmpErrs, fieldIsEmpty(string(acct.Password), "password"))

	tmpErrs = append(tmpErrs, fieldHasPunct(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasPunct(acct.Lname, "lname"))

	tmpErrs = append(tmpErrs, fieldHasNumbers(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasNumbers(acct.Lname, "lname"))

	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Fname, "fname"))
	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Lname, "lname"))
	tmpErrs = append(tmpErrs, fieldHasSymbols(acct.Address, "address"))
	tmpErrs = append(tmpErrs, vetEmailAddress(acct.Email)...)
	tmpErrs = append(tmpErrs, vetPassword(string(acct.Password))...)

	// symbolsFilter := "!@$_^%&*();/-+=\"'`~[]{}<|>"
	// errHasSymbols := errors.New("field can't contain any symbols")
	// if bytes.ContainsAny(acct.Password, symbolsFilter) {
	// 	tmpErrs = append(tmpErrs, fmt.Errorf("error:password:%w", errHasSymbols))
	// }

	for _, err := range tmpErrs {
		if err != nil {
			rsltErr = append(rsltErr, err)
		}
	}

	return tmpErrs
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

// returns a slice of all the emails in the database
func readEmails() []string {
	ctx, db := backend.Connect()
	defer db.Close()

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
func CompareUserCreds(email string, pswd string) error {
	lstEmails := readEmails()

	emailIsReal := false
	for _, target := range lstEmails {
		if email == target {
			emailIsReal = true
			break
		}
	}

	if !emailIsReal {
		errEmailNotFound := errors.New("error: email does not exist")
		log.Println(errEmailNotFound)
		return errEmailNotFound
	}

	// need pswd hash from database
	// compare to one in database

	ctx, db := backend.Connect()
	defer db.Close()

	if len(pswd) < 15 {
		errPswdTooShort := errors.New("error: password is too short, password should be at minimum 15 or more")
		log.Println(errPswdTooShort)
		return errPswdTooShort
	}

	var users []*UserAccount
	err := pgxscan.Select(ctx, db, &users, `Select * FROM user_account`)

	if err != nil {
		errFailedDBEntry := errors.New("database error: couldn’t connect to drumstick")
		log.Println(errFailedDBEntry)
		return errFailedDBEntry
	}

	if len(users) == 0 {
		errNoRowFound := errors.New("query error: user accounts are empty")
		log.Println(errNoRowFound)
		return errNoRowFound
	}

	hashIsReal := false
	for _, user := range users {
		if bcrypt.CompareHashAndPassword(user.Password, []byte(pswd)) == nil {
			hashIsReal = true
			break
		}
	}

	if !hashIsReal {
		errMismatchedHash := errors.New("password mismatch error: hash does not match password")
		log.Println(errMismatchedHash)
		return errMismatchedHash
	}

	return nil
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

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errors.New("error: resource not found: id does not exist")
	}

	if err != nil {
		return 0, fmt.Errorf("error: %w", err)
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
