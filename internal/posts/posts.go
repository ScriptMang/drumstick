package posts

import (
	"errors"
	"fmt"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/backend"

	"github.com/jackc/pgx/v5"
)

type Post struct {
	ID            int    `json:"id" form:"id"`
	UserID        int    `json:"user_id" form:"user_id"`
	Sender        string `json:"sender" form:"sender"`
	Receiver      string `json:"receiver" form:"receiver"`
	Content       string `json:"content" form:"content"`
	NumbComments  int    `json:"number_comments" form:"number_comments"`
	NumbReposts   int    `json:"number_reposts" form:"number_reposts"`
	NumbLikes     int    `json:"number_likes" form:"number_likes"`
	NumbViews     int    `json:"number_views" form:"number_views"`
	NumbBookmarks int    `json:"number_bookmarks" form:"number_bookmarks"`
}

func CreatePosts(userPost, email string) ([]*Post, error) {
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
	var tempPost []*Post
	err := db.QueryRow(ctx, `INSERT ONTO posts(
            user_id, sender,
            receiver, content,
            number_comments, number_reposts,
            number_likes, number_views, number_bookmarks 
    ) VALUES($1,$2,$3,$4,$5,$6,$7)`, uid, userPost, 0, 0, 0, 0, 0).Scan(&tempPost)

	// check list all the errs
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return tempPost, nil
}
