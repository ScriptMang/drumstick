package accts

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"scriptmang/drumstick/internal/templateRenderer"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupEchoClient() *echo.Echo {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.ParseGlob("../../ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

	return r
}

func printAcctErrs(attribNames []string, errMsgs []error) []error {
	var tempErrLst []error
	for i, errMsg := range errMsgs {
		tempErrLst = append(tempErrLst, fmt.Errorf("error:%s:%w", attribNames[i], errMsg))
	}
	return tempErrLst
}

func Test_encryptPassword(t *testing.T) {

	longPswd := bcrypt.ErrPasswordTooLong
	samplePswd1 := []byte("superfragilistic alidocious superfragilistic alidocious superfragilistic alidocious superfragilistic alidocious")
	samplePswd2 := []byte("TheRightAmount2")

	_, actualErr1 := encryptPassword(samplePswd1)
	assert.EqualError(t, actualErr1, longPswd.Error())

	_, actualErr2 := encryptPassword(samplePswd2)
	if actualErr2 != nil {
		t.Fatalf("error:pswd:%s", actualErr2.Error())
	}
}

func Test_accountcreation(t *testing.T) {
	r := setupEchoClient()

	// field errors
	errEmptyField := errors.New("field is empty")
	errHasNums := errors.New("field can't contain any numbers")
	errHasSymbols := errors.New("field can't contain any symbols")
	errHasPunct := errors.New("field has punctuation")

	// email errs
	errReqNums := errors.New("email requires numbers.")
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// password errs
	emptyPswd := errors.New("field can't be empty")
	shortPswd := errors.New("field too short")
	longPswd := errors.New("field can't be long")
	pswdHasNoDigits := errors.New("field missing at least 1 digit")
	missingCapitalLetter := errors.New("field is missing at least 1 capital letter")

	tests := []struct {
		testName                     string
		attribNames                  []string
		fname, lname, address, email string
		password                     []byte
		errLst                       []error
	}{
		{"Empty Attributes", []string{"fname", "lname"}, "", "", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errEmptyField, errEmptyField}},
		{"No Spaces Found in Fname", []string{"fname"}, " Jon", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasPunct}},
		{"No Spaces Found in Lname", []string{"lname"}, "Jon", " Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasPunct}},
		{"No Spaces Found in Fname and Lname", []string{"fname", "lname"}, " Jon", " Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasPunct, errHasPunct}},
		{"No Symbols in Fname", []string{"fname"}, "Jon@", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasSymbols}},
		{"No Symbols in Lname", []string{"lname"}, "Jon", "@Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasSymbols}},
		{"No Numbers in Fname", []string{"fname"}, "Jon3", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasNums}},
		{"No Numbers in Lname", []string{"lname"}, "Jon", "Martin5", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasNums}},
		{"No Symbols in Address", []string{"address"}, "Jon", "Martin", "@409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp432"), []error{errHasSymbols}},
		{"Missing '@' Symbol in Email", []string{"email"}, "Jon", "Martin", "409 Alistar Road", "dummy59gmail.com", []byte("passwordPomp432"), []error{errReqSymbol}},
		{"Missing digits in Email", []string{"email", "email"}, "Jon", "Martin", "409 Alistar Road", "dummy@gmail.com", []byte("passwordPomp432"), []error{errReqNums}},
		{"Missing digits and Invalid ending address in Email", []string{"email", "email"}, "Jon", "Martin", "409 Alistar Road", "dummy@gmail.dum", []byte("passwordPomp432"), []error{errReqNums, errReqEndingAddr}},
		{"Missing at least One Capital Letter in Password", []string{"password"}, "Jon", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordpomp432"), []error{missingCapitalLetter}},
		{"Password is empty", []string{"password"}, "Jon", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte(""), []error{emptyPswd}},
		{"Password is too short", []string{"password"}, "Jon", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("password"), []error{shortPswd, missingCapitalLetter, pswdHasNoDigits}},
		{"Password is too long", []string{"password"}, "Jon", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordPomp43233"), []error{longPswd}},
		{"Password has no digits", []string{"password"}, "Jon", "Martin", "409 Alistar Road", "dummy59@gmail.com", []byte("passwordpompp"), []error{longPswd, missingCapitalLetter, pswdHasNoDigits}},
	}

	// range over table tests and validate the right errs are being thrown
	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/view", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			tempAcct := Account{Fname: tt.fname, Lname: tt.lname, Address: tt.address, Email: tt.email, Password: tt.password}
			actualErrLst := VetAllFields(tempAcct)
			expectedErrLst := printAcctErrs(tt.attribNames, tt.errLst)
			assert.Equal(t, expectedErrLst, actualErrLst, expectedErrLst)

		})
	}
}

func Test_userlogin(t *testing.T) {
	r := setupEchoClient()

	invalidUserCreds := errors.New("incorrect email or password.")
	failedDBConn := errors.New("database error: couldn’t connect to drumstick")
	tests := []struct {
		testName string
		username string
		password string
	}{
		{"Empty values", "", ""},
		{"bad password", "dummy2", "dummy54Dwdgfdgf"},
		{"bad username", "dummy2", "mldKMuffinDrop4"},
		{"password is too long", "dummy2", "lkdrWkkkkkkkkk"},
		{"password is too short", "dummy2", "dfdf"},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/posts", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			actualErr := CompareUserCreds(tt.username, tt.password)

			if errors.Is(actualErr, invalidUserCreds) {
				assert.Equal(t, invalidUserCreds, actualErr, invalidUserCreds)
			}

			if errors.Is(actualErr, failedDBConn) {
				assert.EqualError(t, actualErr, failedDBConn.Error())
				return
			}
		})
	}
}

// test userid retrieval by email
func Test_UserIDByEmail(t *testing.T) {
	rscNotFound := errors.New("error: resource not found: id does not exist")
	sample := "dummy@gmail.com"
	actual, actualErr := UserIDByEmail(sample)

	if errors.Is(actualErr, pgx.ErrNoRows) {
		assert.EqualError(t, actualErr, pgx.ErrNoRows.Error())
	}

	if errors.Is(actualErr, rscNotFound) {
		assert.EqualError(t, actualErr, rscNotFound.Error())
	}

	if actualErr == nil {
		assert.Equal(t, 1, actual)
	}
}
