package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"scriptmang/drumstick/internal/accts"

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

// func viewPosts(w http.ResponseWriter, r *http.Request) {
// }

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
	newAcct.Username = c.FormValue("username")
	newAcct.Password = []byte(c.FormValue("password"))

	rsltErr := accts.VetEmptyFields(newAcct)
	if len(rsltErr) > 0 {
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

func login(c echo.Context) error {
	str := "Login to Drumstick"
	return c.Render(http.StatusOK, "login", str)
}

func vetLogin(c echo.Context) error {
	usr := c.FormValue("username")
	pswd := c.FormValue("password")

	rsltErr := accts.VetUserCreds(usr, pswd)
	if len(rsltErr) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, errors.Join(rsltErr...))
	}

	return c.Render(http.StatusOK, "posts", "My Feed")
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
	router.GET("/login", login)
	router.POST("/view", accountCreation)
	router.POST("/posts", vetLogin)
	router.Logger.Fatal(router.Start(":8080"))
}
