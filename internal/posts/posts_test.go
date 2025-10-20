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
		postMsgs := []string{"message 1", "message 2"}
		var actualPosts []*Post
		for _, msg := range postMsgs {
			testPost, postErr := CreatePosts(ctx, tx, msg, dummyAcct.Email)
			if postErr != nil {
				t.Fatal(postErr)
			}
			actualPosts = append(actualPosts, testPost)
		}

		dummyPosts, err := UserPostsByUserID(ctx, tx, uid)
		if err != nil {
			t.Fatal(err)
		}

		for idx, dummyPost := range dummyPosts {
			if dummyPost.Content != actualPosts[idx].Content {
				t.Fatalf("Expected: %s but got: %s", actualPosts[idx].Content, dummyPost.Content)
			}
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

func Test_UserPostByID(t *testing.T) {
	pool := testutils.TestPool(t)
	defer pool.Close()

	testutils.ResetAndTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
		dummyAcct := accts.Account{
			Fname:    "bob",
			Lname:    "anglefish",
			Username: "tester1",
			Address:  "174 maple street",
			Email:    "lazors504@gmail.com",
			Password: []byte("crazyMango003"),
		}

		_, regErr := accts.CreateAcct(ctx, tx, dummyAcct)
		if regErr != nil {
			t.Fatal(regErr)
		}

		postMsgs := []string{"message 1", "message 2"}
		var expectedPosts []*Post
		for _, msg := range postMsgs {
			testPost, postErr := CreatePosts(ctx, tx, msg, dummyAcct.Email)
			if postErr != nil {
				t.Fatal(postErr)
			}
			expectedPosts = append(expectedPosts, testPost)
		}

		actualPost, err := UserPostByID(ctx, tx, 1)
		if err != nil {
			t.Fatal(err)
		}

		if actualPost.UserID != expectedPosts[0].UserID {
			t.Fatalf("Expected: %d but got: %d", expectedPosts[0].UserID, actualPost.UserID)

func Test_UserIDByPostID(t *testing.T) {
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

		var testPosts []*Post
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
			testPosts = append(testPosts, testPost)
		}

		for i := 0; i < 2; i++ {
			val, err := UserIDByPostID(ctx, tx, i+1)
			if err != nil {
				t.Fatal(err)
			}
			expectedUserID := testPosts[i].UserID
			if val != expectedUserID {
				t.Fatalf("Expected: %d but got: %d", expectedUserID, val)
			}
		}
	})
}
