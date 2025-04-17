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
		testName, attribToBeTested, fname, lname, address, email string
		password                                                 []byte

		// expected                               string
	}{
		{"Empty Attributes", "fname", "", "", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Spaces Found in Fname", "fname", " Jon", " Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Spaces Found in Lname", "lname", "Jon", " Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Spaces Found in Fname and Lname", "fname,lname", " Jon", " Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Symbols in Fname", "fname", "Jon@", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Symbols in Lname", "lname", "Jon", "@Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Numbers in Fname", "fname", "Jon3", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Numbers in Lname", "lname", "Jon", " Martin5", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"No Symbols in Address", "address", "Jon", "Martin", "@409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"Missing '@' Symbol in Email", "email", "Jon", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertStorm432")},
		{"Have at least One Capital Letter in Password", "password", "Jon", "Martin", "409 Alistar Road", "jnp59@gmail.com", []byte("desertstorm432")},
	}

	// need table tests`

	// empty field error
	errEmptyField := errors.New("field is empty")

	// email errs
	errReqNums := errors.New("email requires numbers.")
	errReqSymbol := errors.New("email is missing an '@' symbol.")
	errReqEndingAddr := errors.New("email doesn't match any of the ending addresses.")

	// errs are placed together in json separated by '/n'
	for _, tt := range tests {
		req, _ := http.NewRequest("POST", "/view", nil)
		w := httptest.NewRecorder()
		c := r.NewContext(req, w)
		c.SetPath("/posts")
		t.Run(tt.testName, func(t *testing.T) {
			tempAcct := accts.Account{Fname: tt.fname, Lname: tt.lname, Address: tt.address, Email: tt.email, Password: tt.password}
			errLsts := accts.VetAllFields(tempAcct)
			for _, tgtErr := range errLsts {
				assert.EqualError(t, tgtErr, fmt.Sprintf("error:%s:%s", tt.attribToBeTested, errEmptyField.Error()))
				assert.EqualError(t, tgtErr, errReqNums.Error())
				assert.EqualError(t, tgtErr, errReqEndingAddr.Error())
				assert.EqualError(t, tgtErr, errReqSymbol.Error())
			}
		})
	}

}
