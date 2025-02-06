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
)

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


	ctx, db := backend.CreatePool()
	defer db.Close()

	var userProf accts.UserProfile
	err := pgxscan.Select(ctx, db, &userProf, `SELECT * FROM user_profile`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		os.Exit(1)
func homePage(c echo.Context) error {
	data := "Welcome to drumstick"
	return c.Render(http.StatusOK, "home", data)
}

// func login(w http.ResponseWriter, r *http.Request) {

// }

func main() {
	tm := &TemplateManager{
		templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}
	router := echo.New()
	router.Renderer = tm
	router.GET("/", homePage)
	router.GET("/signup", signUp)
	router.Logger.Fatal(router.Start(":8080"))
}
