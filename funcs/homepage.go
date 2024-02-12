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

	var P []Post

	if r.Method == "GET" {
		database, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatalln("Error opening database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer database.Close()

		err = displaydata(database, &P)
		if err != nil {
			log.Fatalln("Error fetching data from database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = hometmp.Execute(w, P)
		if err != nil {
			log.Fatalln("Error executing template:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		// Post creating code here
	}
}

func insertPost(db *sql.DB, id, post, creator, date string, likes, dislikes int) error {
	insertapost := `INSERT INTO POST(POSTID, CONTENT, CREATORID, LIKES, DISLIKES, DATE) VALUES(?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertapost)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer statement.Close()
	fmt.Println(id, post, creator, likes, dislikes, date)
	_, err = statement.Exec(id, post, creator, likes, dislikes, date)
	if err != nil {
		return fmt.Errorf("error executing statement: %v", err)
	}

	return nil
}

func displaydata(db *sql.DB, P *[]Post) error {
	rows, err := db.Query("SELECT * FROM POST")
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        string
			posttext  string
			creatorid string
			likes     int
			dislikes  int
			data      string
		)

		err := rows.Scan(&id, &posttext, &creatorid, &likes, &dislikes, &data)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		newPost := Post{
			ID:   id,
			Text: posttext,
		}
		*P = append(*P, newPost)
	}

	return nil
}
