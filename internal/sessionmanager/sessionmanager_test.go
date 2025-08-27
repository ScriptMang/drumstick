package sessionmanager

import (
	"html/template"
	"os"
	"scriptmang/drumstick/internal/templateRenderer"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func setupEchoClient() *echo.Echo {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.ParseGlob("../../ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm
	return r
}

func Test_invalidClaimsType(t *testing.T) {
	secret := []byte(os.Getenv("HMAC_SECRET"))
	expirDate := time.Now().Add(time.Hour * 72).Unix()
	claims := &UserCustomClaims{
		Email:      "dummy@gmail.com",
		IsLoggedIn: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirDate, 0)),
		},
	}

	tokenStr, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	token, err := jwt.ParseWithClaims(tokenStr, &UserCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err == nil && token.Valid {
		t.Errorf("Expected claims type mismatch error")
	}
}
