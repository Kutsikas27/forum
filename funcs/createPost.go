package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var createPostPage = template.Must(template.New("createPost.html").ParseFiles("frontend/templates/createPost.html"))

func CreatePost(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/createPost" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		err := createPostPage.Execute(w, nil)
		if err != nil {
			log.Fatal("Error executing template:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "POST" {
		token := true
		User := ""

		c, err := r.Cookie("session_token")
		if err != nil {
			fmt.Println("noo cookie :(")
			token = false
		}

		if token {
			sessionToken := c.Value
			fmt.Println("COOKIE >:D")

			userSession := sessions[sessionToken]
			if userSession.isExpired() {
				delete(sessions, sessionToken)
			} else {
				userSession.expiry = time.Now().Add(120 * time.Second)
			}
			User = userSession.UserData
			fmt.Println(userSession.UserData)
		}
		if User == "" {
			fmt.Println("create user to post")
			return
		}
		r.ParseForm()
		title := r.Form.Get("title")
		text := r.Form.Get("text")
		category := r.Form.Get("category")

		insertPostIntoDB(title, text, category, User)
		http.Redirect(w, r, "/", http.StatusFound)
	}

}

// puts newly created post into the database
func insertPostIntoDB(title, text, category, user string) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO post(title, text, category, creator) VALUES (?, ?, ?, ?)", title, text, category, user)
	if err != nil {
		log.Fatal(err)
	}
}
