package sessionmanager

import (
	"errors"
	"fmt"
	"net/http"
	"scriptmang/drumstick/internal/backend"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Session struct {
	Token  string
	Data   string
	Expiry time.Time
}

type UserCustomClaims struct {
	Email      string `json:"email"`
	IsLoggedIn bool   `json:"isloggedin"`
	jwt.RegisteredClaims
}

func Email(c echo.Context) (string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", echo.NewHTTPError(http.StatusBadRequest, "error:sessionamanager: failed to convert to jwt token")
	}
	claims, ok := user.Claims.(*UserCustomClaims)
	if !ok {
		return "", echo.NewHTTPError(http.StatusBadRequest, "error:sessionamanager: failed to convert to customs claims type")
	}
	email := claims.Email
	if email == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "failed to convert email")
	}
	return email, nil
}

// adds user token to sessions table in database
func CreateToken(usr string) (string, error) {
	expirDate := time.Now().Add(time.Hour * 72).Unix()
	claims := &UserCustomClaims{
		Email:      usr,
		IsLoggedIn: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirDate, 0)),
		},
	}

	fmt.Printf("foo: %v\n", claims.Email)

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return t, nil
}

// adds user token to sessions table in database
func AddToken(tk *jwt.Token) error {
	ctx, db := backend.Connect()
	defer db.Close()

	// need to move custom claims to its own pkg
	claims, ok := tk.Claims.(UserCustomClaims)
	if !ok {
		return errors.New("error: failed conversion for custom claim")
	}

	expTime, err := tk.Claims.GetExpirationTime()

	// err retrieving expiration time
	if err != nil {
		return errors.New("error: failed to get expiration time")
	}

	// insert token
	_, err = db.Exec(
		ctx,
		`INSERT INTO sessions (token, data, expiry) VALUES($1, $2, $3)`,
		tk, claims.Email, expTime.Time,
	)

	// handle multiple errs when insertion goes wrong
	if err != nil {
		return fmt.Errorf("error: failed to add usersession: %w", err)
	}
	return nil
}
