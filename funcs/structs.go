package funcs

import "database/sql"

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
	User     User
}

type PostUser struct {
	Post []Post
	User User
}

type Comment struct {
	PostId  string
	Id      string
	Text    string
	Creator string
	Date    string
}

type Categorys struct {
	Category []string
}

func (p *Post) GetLikes(postId string) int {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return 0
	}
	defer db.Close()
	stmt := `SELECT postid FROM likes WHERE postid = ?`
	rows, err := db.Query(stmt, postId)
	if err != nil {
		return 0
	}
	defer rows.Close()
	likes := 0
	for rows.Next() {
		likes++
	}
	return likes
}

func (p *Post) GetDislikes(postId string) int {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return 0
	}
	defer db.Close()
	stmt := `SELECT postid FROM dislikes WHERE postid = ?`
	rows, err := db.Query(stmt, postId)
	if err != nil {
		return 0
	}
	defer rows.Close()
	dislikes := 0
	for rows.Next() {
		dislikes++
	}
	return dislikes
}
