package funcs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var hometmp = template.Must(template.New("index.html").ParseFiles("web/templates/index.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		HandleErrorPage(w, r)
		return
	}

	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var userName string
	var useruuid string
	var sessionToken string
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken = cookie.Value
		fmt.Println("COOKIE >:D")
		fmt.Println(sessions)
		for index, session := range sessions {
			fmt.Println("session: ", session, "token: ", sessionToken)
			if session.isExpired() || session.UserName == "" {
				fmt.Println("EXPIRED SESSION")
				copy(sessions[index:], sessions[index+1:]) // Shift sessions to fill the gap
				sessions = sessions[:len(sessions)-1]      // Reduce the length of the slice
				deletedCookie := http.Cookie{
					Name:    "session_token",
					Value:   "",
					Expires: time.Unix(0, 0),
				}
				http.SetCookie(w, &deletedCookie)
			} else if session.UserUUID == sessionToken {
				fmt.Println("SESSION FOUND")
				useruuid = session.UserUUID
				userName = session.UserName
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
		if err := initializeTable(); err != nil {
			log.Fatal(err)
		}
		var PU PostUser
		var U User
		posts, err := fetchPostsFromDB()
		if err != nil {
			log.Fatal("Error fetching posts from DB:", err)
			return
		}
		U = User{UserName: userName}
		PU = PostUser{posts, U}
		if err := hometmp.Execute(w, PU); err != nil {
			log.Fatal("Error executing home template:", err)
			return
		}
	} else if r.Method == "POST" {
		if useruuid == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		likes := r.FormValue("like")
		dislikes := r.FormValue("dislike")
		if err := HandleVotes(likes, dislikes, useruuid, w); err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func fetchPostsFromDB() ([]Post, error) {
	var posts []Post

	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, text, category, creator, uuid FROM post")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Text, &post.Category, &post.Creator, &post.Uuid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	reversedPosts := make([]Post, len(posts))

	// make newer posts appear on top of screen
	lastIndex := len(posts) - 1
	for i, post := range posts {
		reversedPosts[lastIndex-i] = post
	}

	return reversedPosts, nil
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func HandleVotes(likes, dislikes, useruuid string, w http.ResponseWriter) error {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	if likes != "" {
		exists1, err := checkLikes(db, likes, useruuid)
		if err != nil {
			return err
		}
		exists2, err := checkDislikes(db, likes, useruuid)
		if err != nil {
			return err
		}
		if !exists1 && !exists2 {
			stmt := "INSERT INTO likes (postid, userid) VALUES (?, ?)"
			_, err = db.Exec(stmt, likes, useruuid)
			if err != nil {
				return err
			}
		} else if exists2 {
			stmt := "DELETE FROM dislikes WHERE postid = ? AND userid = ?"
			_, err = db.Exec(stmt, likes, useruuid)
			if err != nil {
				return err
			}
			stmt = "INSERT INTO likes (postid, userid) VALUES (?, ?)"
			_, err = db.Exec(stmt, likes, useruuid)
			if err != nil {
				return err
			}
		} else if exists1 {
			stmt := "DELETE FROM likes WHERE postid = ? AND userid = ?"
			_, err = db.Exec(stmt, likes, useruuid)
			if err != nil {
				return err
			}
		}
	} else if dislikes != "" {
		exists1, err := checkLikes(db, dislikes, useruuid)
		if err != nil {
			return err
		}
		exists2, err := checkDislikes(db, dislikes, useruuid)
		if err != nil {
			return err
		}
		if !exists1 && !exists2 {
			stmt := "INSERT INTO dislikes (postid, userid) VALUES (?, ?)"
			_, err = db.Exec(stmt, dislikes, useruuid)
			if err != nil {
				return err
			}
		} else if exists1 {
			stmt := "DELETE FROM likes WHERE postid = ? AND userid = ?"
			_, err = db.Exec(stmt, dislikes, useruuid)
			if err != nil {
				return err
			}
			stmt = "INSERT INTO dislikes (postid, userid) VALUES (?, ?)"
			_, err = db.Exec(stmt, dislikes, useruuid)
			if err != nil {
				return err
			}
		} else if exists2 {
			stmt := "DELETE FROM dislikes WHERE postid = ? AND userid = ?"
			_, err = db.Exec(stmt, dislikes, useruuid)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
