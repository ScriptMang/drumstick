package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"

	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/posts"
	"scriptmang/drumstick/internal/sessionmanager"
	"scriptmang/drumstick/internal/templateRenderer"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func signUp(c echo.Context) error {
	data := "Register a New User"
	return c.Render(http.StatusOK, "signup", data)
}

func accountCreation(c echo.Context) error {
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

	_, err := accts.CreateAcct(newAcct)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	//  acct is created -> create a token
	tk, err := sessionmanager.CreateToken(newAcct.Email, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	t, err := tk.SignedString([]byte(os.Getenv("HMAC_SECRET")))
	if err != nil {
		return err
	}

	expTime, expTimeErr := tk.Claims.GetExpirationTime()
	if expTimeErr != nil {
		return errors.New("error: failed to get expiration time")
	}

	ck := &http.Cookie{
		Name:     "auth",
		Value:    t,
		Quoted:   false,
		Expires:  expTime.Time,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(ck)
	c.Request().AddCookie(ck)

	// fmt.Println(resp)
	return c.Render(http.StatusOK, "posts", "Add Posts to Your Feed")
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

	tk, err := sessionmanager.CreateToken(usr, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	t, err := tk.SignedString([]byte(os.Getenv("HMAC_SECRET")))
	if err != nil {
		return err
	}

	expTime, expTimeErr := tk.Claims.GetExpirationTime()
	if expTimeErr != nil {
		return errors.New("error: failed to get expiration time")
	}

	ck := &http.Cookie{
		Name:     "auth",
		Value:    t,
		Quoted:   false,
		Expires:  expTime.Time,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(ck)
	c.Request().AddCookie(ck)
	return c.Render(http.StatusOK, "posts", "Add Posts to Your Feed")
}

// this funct is only  for testing purpsos
func restricted(c echo.Context) error {
	t, err := c.Cookie("auth")
	if err != nil {
		return c.HTML(http.StatusUnauthorized,
			"No authorization token",
		)
	}

	claims, err := sessionmanager.GetUserCustomClaims(t.Value, []byte(os.Getenv("HMAC_SECRET")))

	// returns an http error on failure
	if errors.Is(err, jwt.ErrTokenExpired) {
		return errors.New("token error: token Expired")
	}

	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return errors.New("token error: token has an invalid signature")
	}

	if errors.Is(err, jwt.ErrTokenRequiredClaimMissing) {
		return errors.New("token error: token is missing required claims")
	}

	if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		return errors.New("token error: token signature is invalid")
	}

	email := claims.Email
	return c.Render(http.StatusOK, "sample", "Welcome "+email+"!")
}

func addPosts(c echo.Context) error {
	myPost := c.FormValue("content")

	token, cookieErr := c.Cookie("auth")
	if cookieErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cookieErr)
	}

	claims, retrievalErr := sessionmanager.GetUserCustomClaims(token.Value, []byte(os.Getenv("HMAC_SECRET")))

	if retrievalErr != nil {
		return retrievalErr
	}

	newPost, err := posts.CreatePosts(myPost, claims.Email)

	// handle errs where creating a post fails
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// pass list of the user posts
	// to their posts page

	return c.Render(http.StatusOK, "refresh", newPost.Content)
}

func main() {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	router := echo.New()

	router.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	router.Renderer = tm

	router.GET("/", homePage)
	router.GET("/signup", signUp)
	router.GET("/loginForm", loginForm)
	router.POST("/view", accountCreation)
	router.POST("/posts", vetLogin)
	router.POST("/refresh", addPosts)

	r := router.Group("/restricted")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(sessionmanager.UserCustomClaims)
		},
		SigningKey:  []byte(os.Getenv("HMAC_SECRET")),
		TokenLookup: "cookie:auth",
	}
	r.Use(echojwt.WithConfig(config))

	// Restricted group
	r.GET("", restricted)

	router.Logger.Fatal(router.Start(":8080"))
}
