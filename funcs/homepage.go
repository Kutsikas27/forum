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
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	token := true

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
		fmt.Println(userSession.UserData)
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

	rows, err := db.Query("SELECT id, title, text, category FROM post")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Text, &post.Category)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, post)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
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
		category TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}
