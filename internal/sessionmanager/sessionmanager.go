package sessionmanager

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type UserCustomClaims struct {
	Email      string `json:"email"`
	IsLoggedIn bool   `json:"isloggedin"`
	jwt.RegisteredClaims
}

func GetUserCustomClaims(t string, secretKey []byte) (*UserCustomClaims, error) {

	token, err := jwt.ParseWithClaims(t, &UserCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(*UserCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")

}

func GetEmail(c echo.Context) (string, error) {
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
func CreateToken(usr string, c echo.Context) (*jwt.Token, error) {
	expirDate := time.Now().Add(time.Hour * 72).Unix()
	claims := &UserCustomClaims{
		Email:      usr,
		IsLoggedIn: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirDate, 0)),
		},
	}

	// fmt.Printf("foo: %v\n", claims.Email)

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token, nil
}
