package accts

type UserProfile struct {
	id      int    `json:"id" form:"id"`
	user_id int    `json:"user_id" form:"user_id"`
	fname   string `json:"fname" form:"fname"`
	lname   string `json:"lname" form:"lname"`
	address string `json:"address" form:"address"`
}

type UserAccount struct {
	id       int    `json:"id" form:"id"`
	username string `json:"username" form:"username"`
	password []byte `json:"password" form:"password"`
}

type Posts struct {
	id               int    `json:"id" form:"id"`
	user_id          int    `json:"user_id" form:"user_id"`
	content          string `json:"content" form:"content"`
	number_comments  int    `json:"number_comments" form:"number_comments"`
	number_reposts   int    `json:"number_reposts" form:"number_reposts"`
	number_likes     int    `json:"number_likes" form:"number_likes"`
	number_views     int    `json:"number_views" form:"number_views"`
	number_bookmarks int    `json:"number_bookmarks" form:"number_bookmarks"`
}
