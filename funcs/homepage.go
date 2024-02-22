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

var hometmp = template.Must(template.New("index.html").ParseFiles("frontend/templates/index.html"))

func Homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		HandleErrorPage(w, r)
		return
	}

	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var sessionToken string
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken = cookie.Value
		fmt.Println("COOKIE >:D")

		userSession, exists := sessions[sessionToken]
		if !exists || userSession.isExpired() {
			fmt.Println(exists, userSession.isExpired())
			fmt.Println("why cookey no work :(")
			delete(sessions, sessionToken)

			// Delete the cookie by setting an expired cookie
			deletedCookie := http.Cookie{
				Name:    "session_token",
				Value:   "",
				Expires: time.Unix(0, 0),
			}
			http.SetCookie(w, &deletedCookie)
		} else {
			userSession.expiry = time.Now().Add(120 * time.Second)
			fmt.Println(userSession.UserName)
		}
	} else if err != http.ErrNoCookie {
		fmt.Println("COOKIE >:(")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.Method == "GET" {
		err := initializeTable()
		if err != nil {
			log.Fatal(err)
		}
		posts := fetchPostsFromDB()

		err = hometmp.Execute(w, posts)
		if err != nil {
			log.Fatalln("Error executing template:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func fetchPostsFromDB() []Post {
	var posts []Post

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, text, category, creator, uuid FROM post")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Text, &post.Category, &post.Creator, &post.Uuid)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, post)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(count)
	reversedPosts := make([]Post, len(posts))

	// make newer posts appear on top of screen
	lastIndex := len(posts) - 1
	for i, post := range posts {
		reversedPosts[lastIndex-i] = post
	}

	return reversedPosts
}

func initializeTable() error {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS post (
		id INTEGER PRIMARY KEY,
		title TEXT,
		text TEXT,
		category TEXT,
		creator TEXT,
		uuid TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}
