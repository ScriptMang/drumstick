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

// checks string to see if any char contains
// symbols, punctuation, and digits
// func charFilter() bool {

// }

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

	tests := []struct {
		testName, attribName, testType, fname, lname, address, email string
		password                                                     []byte
	}{
		{"Empty Attributes", "fname", "emptyField", "", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Spaces Found in Fname", "fname", "emptyField", " Jon", " Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Spaces Found in Lname", "lname", "emptyField", "Jon", " Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Spaces Found in Fname and Lname", "fname,lname", "emptyField", " Jon", " Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Symbols in Fname", "fname", "noSymbs", "Jon@", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Symbols in Lname", "lname", "noSymbs", "Jon", "@Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Numbers in Fname", "fname", "noNums", "Jon3", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Numbers in Lname", "lname", "noNums", "Jon", " Martin5", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Symbols in Address", "address", "noSymbs", "Jon", "Martin", "@409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"Missing '@' Symbol in Email", "email", "no@", "Jon", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"Have at least One Capital Letter in Password", "password", "noCap", "Jon", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertstorm432")},
	}

	// field errors
	errEmptyField := errors.New("field is empty")
	errHasNums := errors.New("field can't contain any numbers")
	errHasSymbols := errors.New("field can't contain any symbols")

	// email errs
	errReqNums := errors.New("email requires numbers.")
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// range over table tests and validate the right errs are being thrown
	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/view", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			tempAcct := accts.Account{Fname: tt.fname, Lname: tt.lname, Address: tt.address, Email: tt.email, Password: tt.password}
			errLsts := accts.VetAllFields(tempAcct)
			for _, tgtErr := range errLsts {
				switch tt.testType {
				case "emptyField":
					assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribName, errEmptyField.Error()))
				case "noNums":
					assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribName, errHasNums.Error()))
				case "noSymbs":
					assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribName, errHasSymbols.Error()))
				case "email":
					assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribName, errReqNums.Error()))
					assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribName, errReqEndingAddr.Error()))
					assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribName, errReqSymbol.Error()))
				}

			}
		})
	}

}
