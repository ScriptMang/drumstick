package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"scriptmang/drumstick/internal/accts"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

var testTimeStamp time.Time

func TestMain(m *testing.M) {
	log.Println("Setting up for tests")
	testTimeStamp = time.Now()
	exitVal := m.Run()
	log.Println("Clean up after tests")
	os.Exit(exitVal)
}

func setupEchoClient() *echo.Echo {
	tm := &TemplateManager{
		templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
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

// test to see if the hompage route is acessible
func Test_homepage(t *testing.T) {
	// log.Print("Testing Homepage")
	r := setupEchoClient()
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("error:\n %v\n", err)
	}

	w := httptest.NewRecorder()
	c := r.NewContext(req, w)

	c.SetPath("/")
	if assert.NoError(t, homePage(c)) {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, `{"message":"Not Found"}`, w.Body.String())
	}
}

func Test_signup(t *testing.T) {
	r := setupEchoClient()
	req, err := http.NewRequest("GET", "/signup", nil)

	if err != nil {
		t.Fatalf("error:\n %v\n", err)
	}

	w := httptest.NewRecorder()
	c := r.NewContext(req, w)

	c.SetPath("/signup")
	if assert.NoError(t, signUp(c)) {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, `{"message":"Not Found"}`, w.Body.String())
	}
}

func Test_loginpage(t *testing.T) {
	r := setupEchoClient()
	req, err := http.NewRequest("GET", "/loginForm", nil)

	if err != nil {
		t.Fatalf("error:\n %v\n", err)
	}

	w := httptest.NewRecorder()
	c := r.NewContext(req, w)

	c.SetPath("/loginForm")
	if assert.NoError(t, loginForm(c)) {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEqual(t, `{"message":"Not Found"}`, w.Body.String())
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
	}

	// range over table tests and validate the right errs are being thrown
	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/view", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			tempAcct := accts.Account{Fname: tt.fname, Lname: tt.lname, Address: tt.address, Email: tt.email, Password: tt.password}
			actualErrLst := accts.VetAllFields(tempAcct)
			expectedErrLst := printAcctErrs(tt.attribNames, tt.errLst)
			assert.Equal(t, expectedErrLst, actualErrLst, expectedErrLst)

		})
	}
}

func Test_userlogin(t *testing.T) {
	r := setupEchoClient()

	invalidUserCreds := errors.New("incorrect email or password.")
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
			actualErr := accts.CompareUserCreds(tt.username, tt.password)
			assert.Equal(t, invalidUserCreds, actualErr, invalidUserCreds)

		})
	}

}
