package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"
)

// commenttmp is a template for rendering the comment.html file.
var commenttmp = template.Must(template.New("comment.html").ParseFiles("web/templates/comment.html"))

// PostComment handles the logic for posting a comment.
func ThreadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		log.Fatal("Invalid Method")
	}

	var username string
	var sessionToken string
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken = cookie.Value
		fmt.Println("COOKIE >:D")

		for index, session := range sessions {
			if session.isExpired() {
				fmt.Println("EXPIRED SESSION")
				sessions[index] = sessions[len(sessions)-1]
				sessions = sessions[:len(sessions)-1]
				deletedCookie := http.Cookie{
					Name:    "session_token",
					Value:   "",
					Expires: time.Unix(0, 0),
				}
				http.SetCookie(w, &deletedCookie)
			} else if session.UserUUID == sessionToken {
				fmt.Println("SESSION FOUND")
				username = session.UserName
				session.expiry = time.Now().Add(120 * time.Second)
				break
			}
		}
	} else if err != http.ErrNoCookie {
		fmt.Println("COOKIE >:(")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		postId := r.URL.Query().Get("postId")
		fmt.Println(postId)
		var (
			PC PostComments
			C  = getComments(postId)
			P  = getPost(postId)
			U  = User{UserName: username}
		)
		PC = PostComments{P, C, U}
		commenttmp.Execute(w, PC)
	} else if r.Method == "POST" {
		if username == "" {
			http.Error(w, "Log in to make comment", http.StatusUnauthorized)
			return
		}
		comment := r.FormValue("commenttext")
		if comment == "" {
			fmt.Fprintln(w, "Comment cannot be empty")
			return
		}
		addComment(comment, r.URL.Query().Get("postId"), username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// getPost retrieves a post from the database based on the given postId.
func getPost(postId string) Post {
	var P Post
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	stmt := `SELECT title, text, category, creator, uuid FROM post WHERE uuid = ?`
	err = db.QueryRow(stmt, postId).Scan(&P.Title, &P.Text, &P.Category, &P.Creator, &P.Uuid)
	if err != nil {
		log.Fatal(err)
	}
	return P
}

// getComments retrieves all comments for a given postId from the database.
func getComments(postId string) []Comment {
	var C []Comment
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	stmt := `SELECT REPLYID, CONTENT, CREATORID FROM comment WHERE REPLYID = ?`
	rows, err := db.Query(stmt, postId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var c Comment
		err = rows.Scan(&c.PostId, &c.Text, &c.Creator)
		if err != nil {
			log.Fatal(err)
		}
		C = append(C, c)
	}
	reverseComments := make([]Comment, len(C))
	lastIndex := len(C) - 1
	for i, comment := range C {
		reverseComments[lastIndex-i] = comment
	}
	return reverseComments
}

// addComment adds a new comment to the database for the given postId.
func addComment(comment, postId, username string) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	stmt := `INSERT INTO COMMENT (REPLYID, CONTENT, CREATORID) VALUES (?, ?, ?)`
	_, err = db.Exec(stmt, postId, comment, username)
	if err != nil {
		log.Fatal(err)
	}
}
