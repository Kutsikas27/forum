package funcs

type Post struct {
	ID       int
	Title    string
	Text     string
	Category string
	Creator  string
	Uuid     string
}

type User struct {
	ID       string
	Email    string
	UserName string
}

type PostComments struct {
	Post     Post
	Comments []Comment
}

type Comment struct {
	PostId  string
	Id      string
	Text    string
	Creator string
	// Likes int
	// dislikes int
	Date string
}

type Categorys struct {
	Category []string
}
