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
	var useruuid string
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
			useruuid = userSession.UserUUID
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
	} else if r.Method == "POST" {
		if useruuid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		db, err := sql.Open("sqlite3", "database.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		likes := r.FormValue("like")
		dislieks := r.FormValue("dislike")
		fmt.Println(likes)
		fmt.Println(dislieks)
		if likes != "" {
			exists1, err := checkLikes(db, likes, useruuid)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			exists2, err := checkDislikes(db, likes, useruuid)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !exists1 && !exists2 {
				stmt := "INSERT INTO likes (postid, userid) VALUES (?, ?)"
				_, err = db.Exec(stmt, likes, useruuid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else if exists2 {
				stmt := "DELETE FROM dislikes WHERE postid = ? AND userid = ?"
				_, err = db.Exec(stmt, likes, useruuid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				stmt = "INSERT INTO likes (postid, userid) VALUES (?, ?)"
				_, err = db.Exec(stmt, likes, useruuid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else if exists1 {
				// do nothing
			}
		} else if dislieks != "" {
			exists1, err := checkLikes(db, dislieks, useruuid)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			exists2, err := checkDislikes(db, dislieks, useruuid)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !exists1 && !exists2 {
				stmt := "INSERT INTO dislikes (postid, userid) VALUES (?, ?)"
				_, err = db.Exec(stmt, dislieks, useruuid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else if exists1 {
				stmt := "DELETE FROM likes WHERE postid = ? AND userid = ?"
				_, err = db.Exec(stmt, dislieks, useruuid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				stmt = "INSERT INTO dislikes (postid, userid) VALUES (?, ?)"
				_, err = db.Exec(stmt, dislieks, useruuid)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else if exists2 {
				// do nothing
			}
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

func checkLikes(db *sql.DB, postId, userId string) (bool, error) {
	stmt := "SELECT * FROM likes WHERE postid = ? AND userid = ?"
	err := db.QueryRow(stmt, postId, userId).Scan(&postId, &userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func checkDislikes(db *sql.DB, postId, userId string) (bool, error) {
	stmt := "SELECT * FROM dislikes WHERE postid = ? AND userid = ?"
	err := db.QueryRow(stmt, postId, userId).Scan(&postId, &userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
