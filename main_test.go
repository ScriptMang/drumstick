package main

import (
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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

// test to see if the hompage route is acessible
func Test_homepage(t *testing.T) {
	// log.Print("Testing Homepage")
	tm := &TemplateManager{
		templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

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
	tm := &TemplateManager{
		templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

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
	tm := &TemplateManager{
		templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

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
