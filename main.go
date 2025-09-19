package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

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
	newAcct.Username = c.FormValue("username")
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

	userPosts, qryErr := posts.UserPosts()
	if qryErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, qryErr)
	}

	return c.Render(http.StatusOK, "posts", map[string]any{
		"Posts": userPosts,
	})
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

	return c.Redirect(http.StatusSeeOther, "/posts")
}

func viewFeed(c echo.Context) error {

	t, err := c.Cookie("auth")
	if err != nil {
		return c.HTML(http.StatusUnauthorized,
			"No authorization token",
		)
	}

	claims, err := sessionmanager.GetUserCustomClaims(t.Value, []byte(os.Getenv("HMAC_SECRET")))

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
	uid, uidErr := accts.UserIDByEmail(email)
	if uidErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	username, qryErr := posts.UsernameByID(uid)
	if qryErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, qryErr)
	}

	userPosts, qryErr := posts.UserPosts()
	if qryErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, qryErr)
	}

	return c.Render(http.StatusOK, "posts", map[string]any{
		"Username": username,
		"Posts":    userPosts,
	})
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

func deletePosts(c echo.Context) error {

	strPostID := c.Param("id")
	postID, convErr := strconv.Atoi(strPostID)
	if convErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, convErr)
	}

	deleteQueryErr := posts.DeletePostByID(postID)
	if deleteQueryErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, deleteQueryErr)
	}

	return c.Redirect(http.StatusSeeOther, "/posts")
}

func viewResponsePage(c echo.Context) error {

	t, err := c.Cookie("auth")
	if err != nil {
		return c.HTML(http.StatusUnauthorized,
			"No authorization token",
		)
	}

	claims, err := sessionmanager.GetUserCustomClaims(t.Value, []byte(os.Getenv("HMAC_SECRET")))

	email := claims.Email
	uid, uidErr := accts.UserIDByEmail(email)
	if uidErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	username, qryErr := posts.UsernameByID(uid)
	if qryErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, qryErr)
	}

	strID := c.Param("id")
	postID, convErr := strconv.Atoi(strID)
	if convErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, convErr)
	}

	if qryErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, qryErr)
	}

	parentPost, err := posts.UserPostByID(postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, "reply", map[string]any{
		"ChildPostUsername": username,
		"ParentPost":        parentPost,
	})
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

	_, err := posts.CreatePosts(myPost, claims.Email)

	// handle errs where creating a post fails
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return c.Redirect(http.StatusSeeOther, "/posts")
}

func main() {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.ParseGlob("ui/html/pages/*[^#?!|].tmpl")),
	}

	router := echo.New()

	router.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	router.Static("static", "ui/static/")

	router.Renderer = tm
	router.GET("/", homePage)
	router.GET("/signup", signUp)
	router.POST("/view", accountCreation)
	router.GET("/login", loginForm)
	router.POST("/login", vetLogin)

	router.GET("/posts", viewFeed)
	router.POST("/posts", addPosts)
	router.POST("/posts/:id", deletePosts)
	router.GET("/posts/:id/reply", viewResponsePage)

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
