package accts

type UserProfile struct {
	ID      int    `json:"id" form:"id"`
	User_ID int    `json:"user_id" form:"user_id"`
	Fname   string `json:"fname" form:"fname"`
	Lname   string `json:"lname" form:"lname"`
	Address string `json:"address" form:"address"`
}

type UserAccount struct {
	ID       int    `json:"id" form:"id"`
	Username string `json:"username" form:"username"`
	Password []byte `json:"password" form:"password"`
}

type Posts struct {
	ID               int    `json:"id" form:"id"`
	User_ID          int    `json:"user_id" form:"user_id"`
	Content          string `json:"content" form:"content"`
	Number_Comments  int    `json:"number_comments" form:"number_comments"`
	Number_Reposts   int    `json:"number_reposts" form:"number_reposts"`
	Number_Likes     int    `json:"number_likes" form:"number_likes"`
	Number_Views     int    `json:"number_views" form:"number_views"`
	Number_Bookmarks int    `json:"number_bookmarks" form:"number_bookmarks"`
}
