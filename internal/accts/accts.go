package accts

type UserProfile struct {
	ID      int    `json:"id" form:"id"`
	UserID  int    `json:"user_id" form:"user_id"`
	Fname   string `json:"fname" form:"fname"`
	Lname   string `json:"lname" form:"lname"`
	Address string `json:"address" form:"address"`
}

type UserAccount struct {
	ID       int    `json:"id" form:"id"`
	Username string `json:"username" form:"username"`
	Password []byte `json:"password" form:"password"`
}

type Account struct {
	ID       int    `json:"id" form:"id"`
	Fname    string `json:"fname" form:"fname"`
	Lname    string `json:"lname" form:"lname"`
	Address  string `json:"address" form:"address"`
	Username string `json:"username" form:"username"`
	Password []byte `json:"password" form:"password"`
}

type Posts struct {
	ID            int    `json:"id" form:"id"`
	UserID        int    `json:"user_id" form:"user_id"`
	Content       string `json:"content" form:"content"`
	NumbComments  int    `json:"number_comments" form:"number_comments"`
	NumbReposts   int    `json:"number_reposts" form:"number_reposts"`
	NumbLikes     int    `json:"number_likes" form:"number_likes"`
	NumbViews     int    `json:"number_views" form:"number_views"`
	NumbBookmarks int    `json:"number_bookmarks" form:"number_bookmarks"`
}

}
