package funcs

type Post struct {
	ID       string
	Text     string
	CID      string
	Likes    int
	Dislikes int
	Date     string
}

type User struct {
	ID       string
	Email    string
	UserName string
}

type Comment struct {
	ReplyId  string
	Id       string
	Text     string
	creator  string
	Likes    int
	dislikes int
	Date     string
}

type Categorys struct {
	Category []string
}
