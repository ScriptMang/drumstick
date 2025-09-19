package posts

import (
	"errors"
	"fmt"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/backend"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Post struct {
	ID        int       `json:"id" form:"id"`
	ParentID  int       `json:"parent_id" form:"parent_id"`
	UserID    int       `json:"user_id" form:"user_id"`
	Username  string    `json:"username" form:"username"`
	Content   string    `json:"content" form:"content"`
	CreatedOn time.Time `json:"created_on"`
	Date      string
}

func UsernameByID(id int) (string, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var username string
	err := db.QueryRow(ctx,
		`Select username FROM user_account WHERE id=$1`, id).Scan(&username)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}
	return username, nil
}

func UserPostsByUserID(uid int) ([]*Post, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var userPosts []*Post
	err := pgxscan.Select(ctx, db, &userPosts, `Select * FROM posts WHERE user_id = $1`, uid)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	for _, pst := range userPosts {
		pst.Date = pst.CreatedOn.Format("01/02/2006")
	}

	return userPosts, nil
}

func UserPostByID(postID int) (Post, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var tempPost Post
	queryErr := db.QueryRow(ctx,
		`SELECT id, user_id, username, content FROM posts WHERE id=$1`, postID).
		Scan(&tempPost.ID, &tempPost.UserID, &tempPost.Username, &tempPost.Content)

	if queryErr != nil {
		return tempPost, fmt.Errorf("error: %w", queryErr)
	}

	return tempPost, nil
}

// returns the list of all user posts
func UserPosts() ([]*Post, error) {
	ctx, db := backend.Connect()
	defer db.Close()

	var userPosts []*Post
	err := pgxscan.Select(ctx, db, &userPosts, `Select * FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	for _, pst := range userPosts {
		pst.Date = pst.CreatedOn.Format("01/02/2006")
	}

	return userPosts, nil
}

// deletes the user's post given their post id
// returns the new list of posts
func DeletePostByID(postID int) error {
	ctx, db := backend.Connect()
	defer db.Close()

	_, queryErr := db.Exec(ctx, `DELETE FROM POSTS WHERE id = $1`, postID)
	if queryErr != nil {
		return fmt.Errorf("error:%w", queryErr)
	}

	return nil
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

	// get the  user's  username
	username, userNameErr := UsernameByID(uid)
	if userNameErr != nil {
		return nil, fmt.Errorf("error: %v\n", userNameErr)
	}

	var tempPost = new(Post)
	err := db.QueryRow(ctx, `INSERT INTO posts(parent_id, user_id, username, content)`+
		` VALUES($1,$2,$3,$4) RETURNING *`, 0, uid, username, userPost).Scan(
		&tempPost.ID, &tempPost.UserID, &tempPost.Username,
		&tempPost.Content, &tempPost.CreatedOn,
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
