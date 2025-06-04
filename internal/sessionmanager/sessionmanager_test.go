package sessionmanager

import (
	"html/template"
	"os"
	"scriptmang/drumstick/internal/templateRenderer"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
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

func Test_userCustomClaim(t *testing.T) {

	r := setupEchoClient()
	c := r.NewContext(r.AcquireContext().Request(), r.AcquireContext().Response().Writer)

	invalidKey := jwt.ErrInvalidKey
	invalidKeyType := jwt.ErrInvalidKeyType
	hashUnavailable := jwt.ErrHashUnavailable
	tokenMalformed := jwt.ErrTokenMalformed
	tokenUnverifiable := jwt.ErrTokenUnverifiable
	tokenSignatureInvalid := jwt.ErrTokenSignatureInvalid
	tokenRequiredClaimsMissing := jwt.ErrTokenRequiredClaimMissing
	tokenInvalidAudience := jwt.ErrTokenInvalidAudience
	tokenExpired := jwt.ErrTokenExpired
	tokenUsedBeforeIssued := jwt.ErrTokenUsedBeforeIssued
	tokenInvalidIssuer := jwt.ErrTokenInvalidIssuer
	tokenInvalidSubject := jwt.ErrTokenInvalidSubject
	tokenNotValidYet := jwt.ErrTokenNotValidYet
	tokenInvalidID := jwt.ErrTokenInvalidId
	tokenInvalidClaims := jwt.ErrTokenInvalidClaims
	invalidType := jwt.ErrInvalidKeyType

	sampleToken, _ := CreateToken("dummy@gmail.com", c)
	sampleTokenStr := sampleToken.Raw
	env := []byte(os.Getenv("HMAC_SECRET"))
	tests := []struct {
		testName    string
		tokenStr    string
		secretKey   []byte
		expectedErr error
	}{{"invalidKey", sampleTokenStr, env, invalidKey},
		{"invalidKeyType", sampleTokenStr, env, invalidKeyType},
		{"hashUnavailable", sampleTokenStr, env, hashUnavailable},
		{"tokenMalformed", sampleTokenStr, env, tokenMalformed},
		{"tokenUnverifiable", "x/dddddddd", env, tokenUnverifiable},
		{"tokenSignatureInvalid", "dddddddd68686894959d", env, tokenSignatureInvalid},
		{"tokenRequiredClaimsMissing", "dddddddd", env, tokenRequiredClaimsMissing},
		{"tokenInvalidAudience", "dddddddd", env, tokenInvalidAudience},
		{"tokenExpired", "dddddddd", env, tokenExpired},
		{"tokenUsedBeforeIssued", "dddddddd", env, tokenUsedBeforeIssued},
		{"tokenInvalidIssuer", "dddddddd", env, tokenInvalidIssuer},
		{"tokenInvalidSubject", "dddddddd", env, tokenInvalidSubject},
		{"tokenNotValidYet", "dddddddd", env, tokenNotValidYet},
		{"tokenInvalidID", "dddddddd", env, tokenInvalidID},
		{"tokenInvalidClaims", "dddddddd", env, tokenInvalidClaims},
		{"invalidType", "dddddddd", env, invalidType},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			_, actualErr := GetUserCustomClaims(tc.tokenStr, tc.secretKey)

			switch actualErr {
			case invalidKey:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case invalidKeyType:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case hashUnavailable:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenMalformed:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenUnverifiable:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenSignatureInvalid:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenRequiredClaimsMissing:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenInvalidAudience:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenExpired:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenUsedBeforeIssued:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenInvalidIssuer:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenInvalidSubject:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenNotValidYet:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenInvalidID:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case tokenInvalidClaims:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			case invalidType:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			default:
				assert.EqualError(t, actualErr, tc.expectedErr.Error())
			}
		})
	}
}
