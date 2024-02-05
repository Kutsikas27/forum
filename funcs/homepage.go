package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

var hometmp = template.Must(template.New("index.html").ParseFiles("frontend/templates/index.html"))

func Homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var P []Post
	Id := "asdfd"
	posttext := "this is my first post"
	creator := "asdfa"
	likes := 1
	dislikes := 0
	date := "today"

	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatalln("Error opening database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	err = insertPost(database, Id, posttext, creator, date, likes, dislikes)
	if err != nil {
		log.Fatalln("Error inserting post into database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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
