package posts

import (
	"fmt"
	"html/template"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/templateRenderer"
	"scriptmang/drumstick/internal/testutils"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func dict(vals ...any) (map[string]any, error) {
	if len(vals)%2 != 0 {
		return nil, fmt.Errorf("dictionary requires an even number of arguments")
	}

	m := make(map[string]any, len(vals)/2)
	for i := 0; i < len(vals); i += 2 {
		ky, ok := vals[i].(string)
		if !ok {
			return nil, fmt.Errorf("dictionary keys must be string type")
		}
		m[ky] = vals[i+1]
	}
	return m, nil
}

func setupEchoClient() *echo.Echo {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.New("").Funcs(template.FuncMap{
			"dict": dict,
		}).ParseGlob("../../ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

	return r
}

// tests the func that returns the username of a post given the account id
func Test_UsernameByID(t *testing.T) {
	pool := testutils.TestPool(t)
	ctx := t.Context()
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	// before insert need to encrypt password, but it’s a private func
	// meaning i need to create an acct object
	dummyAcct := accts.Account{
		Fname:    "demo",
		Lname:    "man",
		Username: "tester",
		Address:  "174 maple street",
		Email:    "lazors504@gmail.com",
		Password: []byte("crazyMango003"),
	}

	// register the acct -> user-acct & user-profile
	_, regErr := accts.CreateAcct(ctx, tx, dummyAcct)
	if regErr != nil {
		t.Fatal(regErr)
	}

	username, err := UsernameByID(ctx, tx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if username != "tester" {
		t.Errorf("expected tester, got %s", username)
	}
}
