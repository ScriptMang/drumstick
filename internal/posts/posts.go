package posts

import (
	"errors"
	"fmt"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/backend"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Post struct {
	UserID   int    `json:"user_id" form:"user_id"`
	ParentID int    `json:"parent_id"`
	ThreadID int    `json:"thread_id"`
	Content  string `json:"content" form:"content"`
}

func UserPostsByID(uid int) ([]*Post, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var userPosts []*Post
	err := pgxscan.Select(ctx, db, &userPosts, `Select * FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return userPosts, nil
}

func CreatePosts(userPost, email string) (*Post, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	// need get user id of the poster
	uid, uidErr := accts.UserIDByEmail(email)

	if uidErr != nil {
		if errors.Is(uidErr, pgx.ErrNoRows) {
			return nil, errors.New("error: resource not found: id with specified email couldn't be found")
		}
		return nil, fmt.Errorf("error: %w", uidErr)
	}

	// add posts to posts
	var tempPost = new(Post)
	err := db.QueryRow(ctx, `INSERT INTO posts(user_id, parent_id, content)`+
		` VALUES($1,$2,$3) RETURNING *`, uid, 0, userPost).Scan(
		&tempPost.UserID, &tempPost.ParentID,
		&tempPost.ThreadID, &tempPost.Content,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, fmt.Errorf("error: %s", pgErr.Message)
		}
	}

	// check list all the errs
	return tempPost, nil
}
