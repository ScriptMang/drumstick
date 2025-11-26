package posts

import (
	"context"
	"errors"
	"fmt"
	"scriptmang/drumstick/internal/accts"
	"scriptmang/drumstick/internal/backend"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"
)

type Post struct {
	ID        int       `json:"id" form:"id"`
	ParentID  int       `json:"parent_id" form:"parent_id"`
	UserID    int       `json:"user_id" form:"user_id"`
	Username  string    `json:"username" form:"username"`
	Content   string    `json:"content" form:"content"`
	CreatedOn time.Time `json:"created_on"`
	Replies   []*Post
	Date      string
}

func UsernameByID(ctx context.Context, db backend.Querier, id int) (string, error) {
	var username string
	err := db.QueryRow(ctx,
		`Select username FROM user_account WHERE id=$1`, id).Scan(&username)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}
	return username, nil
}

func UserPostsByUserID(ctx context.Context, db backend.Querier, uid int) ([]*Post, error) {
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

func UserIDByPostID(ctx context.Context, db backend.Querier, pid int) (int, error) {
	var userID int
	err := db.QueryRow(ctx,
		`SELECT user_id FROM POSTS WHERE id=$1`, pid).Scan(&userID)

	if err != nil {
		return 0, fmt.Errorf("error: %w", err)
	}
	return userID, nil
}

func UserPostByID(ctx context.Context, db backend.Querier, postID int) (Post, error) {
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
func UserPosts(ctx context.Context, db backend.Querier) ([]*Post, error) {
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
func DeletePostByID(ctx context.Context, db backend.Querier, postID int) error {
	_, queryErr := db.Exec(ctx, `DELETE FROM POSTS WHERE id = $1`, postID)
	if queryErr != nil {
		return fmt.Errorf("error:%w", queryErr)
	}

	return nil
}

func MakePostsTree(posts []*Post) []*Post {
	postsIndex := make(map[int]*Post)
	var topLevelPosts []*Post

	for _, p := range posts {
		postsIndex[p.ID] = p
	}

	for _, p := range posts {
		if p.ParentID == 0 {
			topLevelPosts = append(topLevelPosts, p)
		} else {
			parent := postsIndex[p.ParentID]
			if parent != nil {
				parent.Replies = append(parent.Replies, p)
			}
		}
	}
	return topLevelPosts
}

func CreatePosts(ctx context.Context, db backend.Querier, userPost, email string) (*Post, error) {
	// need get user id of the poster
	uid, uidErr := accts.UserIDByEmail(ctx, db, email)

	if uidErr != nil {
		return nil, uidErr
	}

	// get the  user's  username
	username, userNameErr := UsernameByID(ctx, db, uid)
	if userNameErr != nil {
		return nil, fmt.Errorf("error: %v\n", userNameErr)
	}

	var tempPost = new(Post)
	err := db.QueryRow(ctx, `INSERT INTO posts(parent_id, user_id, username, content)`+
		` VALUES($1,$2,$3,$4) RETURNING id, user_id, username, content, created_on`, 0, uid, username, userPost).Scan(
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

func ReplyToPost(ctx context.Context, db backend.Querier, parentPID int, postBody, email string) (*Post, error) {
	// need get user id of the poster
	uid, uidErr := accts.UserIDByEmail(ctx, db, email)

	if uidErr != nil {
		return nil, uidErr
	}

	// get the  user's  username
	username, userNameErr := UsernameByID(ctx, db, uid)
	if userNameErr != nil {
		return nil, fmt.Errorf("error: %v\n", userNameErr)
	}

	var tempPost = new(Post)
	err := db.QueryRow(ctx,
		`INSERT INTO posts(parent_id, user_id, username, content)`+
			` VALUES($1,$2,$3,$4) RETURNING *`, parentPID, uid, username, postBody).Scan(
		&tempPost.ID, &tempPost.ParentID, &tempPost.UserID, &tempPost.Username,
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
