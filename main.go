package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/posts"
	"scriptmang/drumstick/internal/sessionmanager"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type respBody struct {
	msg string
}

type TemplateManager struct {
	templates *template.Template
}

func (tm *TemplateManager) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}
	err := tm.templates.ExecuteTemplate(w, name, data)

	if err != nil {
		log.Println("template not found")
	}

	return err
}

func signUp(c echo.Context) error {
	data := "Register a New User"
	return c.Render(http.StatusOK, "signup", data)
}

func accountCreation(c echo.Context) error {
	var resp respBody
	var newAcct accts.Account

	newAcct.Fname = c.FormValue("fname")
	newAcct.Lname = c.FormValue("lname")
	newAcct.Address = c.FormValue("address")
	newAcct.Email = c.FormValue("email")
	newAcct.Password = []byte(c.FormValue("password"))

	rsltErr := accts.VetAllFields(newAcct)
	if len(rsltErr) > 0 {
		c.Logger().Error(rsltErr)
		return echo.NewHTTPError(http.StatusBadRequest, errors.Join(rsltErr...))
	}

	msg, err := accts.CreateAcct(newAcct)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	resp.msg = msg
	// fmt.Println(resp)
	return c.Render(http.StatusOK, "view", resp)
}

func homePage(c echo.Context) error {
	data := "Login or Create an Account"
	return c.Render(http.StatusOK, "home", data)
}

func loginForm(c echo.Context) error {
	str := "Login to Drumstick"
	return c.Render(http.StatusOK, "loginForm", str)
}

func vetLogin(c echo.Context) error {
	usr := c.FormValue("email")
	pswd := c.FormValue("password")

	rsltErr := accts.CompareUserCreds(usr, pswd)
	if rsltErr != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, rsltErr)
	}

	t, err := sessionmanager.CreateToken(usr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return c.Render(http.StatusOK, "posts", echo.Map{
		"token": t,
	})
}

// this funct is only  for testing purpsos
func restricted(c echo.Context) error {
	email, err := sessionmanager.Email(c)

	// returns an http error on failure
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "Welcome "+email+"!")
}

func addPosts(c echo.Context) error {
	myPost := c.FormValue("content")

	email, retrievalErr := sessionmanager.Email(c)
	if retrievalErr != nil {
		return retrievalErr
	}

	newPost, err := posts.CreatePosts(myPost, email)

	// handle errs where creating a post fails
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// pass list of the user posts
	// to their posts page

	return c.Render(http.StatusOK, "refresh", *newPost[0])
}

func main() {
	tm := &TemplateManager{
		templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	router := echo.New()
	router.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))

	router.Renderer = tm

	router.GET("/", homePage)
	router.GET("/signup", signUp)
	router.GET("/loginForm", loginForm)
	router.POST("/view", accountCreation)
	router.POST("/posts", vetLogin)
	router.POST("/refresh", addPosts)

	// Restricted group
	r := router.Group("/restricted")

	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(sessionmanager.UserCustomClaims)
		},
		SigningKey: []byte("secret"),
	}
	r.Use(echojwt.WithConfig(config))
	r.GET("", restricted)

	router.Logger.Fatal(router.Start(":8080"))
}
