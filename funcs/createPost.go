package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "github.com/mattn/go-sqlite3"
)

var createPostPage = template.Must(template.New("createPost.html").ParseFiles("frontend/templates/createPost.html"))

func CreatePost(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/createPost" {
		HandleErrorPage(w, r)
		return
	}

	var username string
	var sessionToken string
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken = cookie.Value
		fmt.Println("COOKIE >:D")

		userSession, exists := sessions[sessionToken]
		if !exists || userSession.isExpired() {
			delete(sessions, sessionToken)
			deletedCookie := http.Cookie{
				Name:    "session_token",
				Value:   "",
				Expires: time.Unix(0, 0),
			}
			http.SetCookie(w, &deletedCookie)
		} else {
			userSession.expiry = time.Now().Add(120 * time.Second)
			username = userSession.UserData
			fmt.Println(userSession.UserData)
		}
	} else if err != http.ErrNoCookie {
		fmt.Println("COOKIE >:(")
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		if username == "" {
			http.Error(w, "Log in to make post", http.StatusUnauthorized)
			return
		}
		r.ParseForm()
		title := r.Form.Get("title")
		text := r.Form.Get("text")
		category := r.Form.Get("category")

		insertPostIntoDB(title, text, category, username)
		http.Redirect(w, r, "/", http.StatusFound)
	}

}

// puts newly created post into the database
func insertPostIntoDB(title, text, category, username string) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}

	postUUID, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	strPostUUID := postUUID.String()
	_, err = db.Exec("INSERT INTO post(title, text, category, creator, uuid) VALUES (?, ?, ?, ?, ?)", title, text, category, username, strPostUUID)
	if err != nil {
		log.Fatal(err)
	}
}
