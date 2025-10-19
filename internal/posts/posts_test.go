package posts

import (
	"context"
	"fmt"
	"html/template"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/templateRenderer"
	"scriptmang/drumstick/internal/testutils"
	"testing"

	"github.com/jackc/pgx/v5"
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
	defer pool.Close()

	testutils.ResetAndTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {

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
	})
}

func Test_CreatePosts(t *testing.T) {
	pool := testutils.TestPool(t)
	defer pool.Close()

	testutils.ResetAndTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
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

		// submit a new post
		expected := "message 1"
		testPost, postErr := CreatePosts(ctx, tx, "message 1", dummyAcct.Email)
		if postErr != nil {
			t.Fatal(postErr)
		}

		if testPost.Content != expected {
			t.Fatalf("Expected: %s got: %s", testPost.Content, expected)
		}
	})
}

func Test_UserPostsByUserID(t *testing.T) {
	pool := testutils.TestPool(t)
	defer pool.Close()

	testutils.ResetAndTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
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

		uid, emailErr := accts.UserIDByEmail(ctx, tx, dummyAcct.Email)
		if emailErr != nil {
			t.Fatal(emailErr)
		}

		// submit a new post
		testPost1, postErr := CreatePosts(ctx, tx, "message 1", dummyAcct.Email)
		if postErr != nil {
			t.Fatal(postErr)
		}

		testPost2, postErr := CreatePosts(ctx, tx, "message 2", dummyAcct.Email)
		if postErr != nil {
			t.Fatal(postErr)
		}

		dummyPosts, err := UserPostsByUserID(ctx, tx, uid)
		if err != nil {
			t.Fatal(err)
		}

		if dummyPosts[0].Content != testPost1.Content {
			t.Fatalf("Expected: %s but got: %s", testPost1.Content, dummyPosts[0].Content)
		}

		if dummyPosts[1].Content != testPost2.Content {
			t.Fatalf("Expected: %s but got: %s", testPost2.Content, dummyPosts[1].Content)
		}
	})
}

func Test_UserPosts(t *testing.T) {
	pool := testutils.TestPool(t)
	defer pool.Close()

	testutils.ResetAndTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
		dummyAccts := []accts.Account{{
			Fname:    "bob",
			Lname:    "anglefish",
			Username: "tester1",
			Address:  "174 maple street",
			Email:    "lazors504@gmail.com",
			Password: []byte("crazyMango003"),
		}, {
			Fname:    "jacob",
			Lname:    "loftwood",
			Username: "tester2",
			Address:  "403 mumble street",
			Email:    "baseballChamp@gmail.com",
			Password: []byte("dynamoPoppler952"),
		},
		}

		var actualPosts []*Post
		for idx, dummyAcct := range dummyAccts {
			_, regErr := accts.CreateAcct(ctx, tx, dummyAcct)
			if regErr != nil {
				t.Fatal(regErr)
			}
			msg := fmt.Sprintf("message %d", idx+1)
			testPost, createPostErr := CreatePosts(ctx, tx, msg, dummyAcct.Email)
			if createPostErr != nil {
				t.Fatal(createPostErr)
			}
			actualPosts = append(actualPosts, testPost)
		}

		postLst, postLstErr := UserPosts(ctx, tx)
		if postLstErr != nil {
			t.Fatal(postLstErr)
		}

		for idx, userPost := range postLst {
			if userPost.UserID != actualPosts[idx].UserID {
				t.Fatalf("Expected: %d but got: %d", actualPosts[idx].UserID, userPost.UserID)
			}

			if userPost.Content != actualPosts[idx].Content {
				t.Fatalf("Expected: %s but got: %s", actualPosts[idx].Content, userPost.Content)
			}
		}
	})

}
