package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/templateRenderer"
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
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

	return r
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
