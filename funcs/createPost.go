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

var createPostPage = template.Must(template.New("createPost.html").ParseFiles("web/templates/createPost.html"))

func TopicHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/create-post" {
		HandleErrorPage(w, r)
		return
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
		var U = User{UserName: username}
		err := createPostPage.Execute(w, U)
		if err != nil {
			log.Fatal("Error executing template:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "POST" {
		if username == "" {
			fmt.Fprintln(w, "Log in to create a post")
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
