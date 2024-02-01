package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID   int
	Text string
}

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
	Id := 1
	posttext := "this is my first post"

	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	err = insertPost(database, Id, posttext)
	if err != nil {
		log.Fatal("Error inserting post into database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = displaydata(database, &P)
	if err != nil {
		log.Fatal("Error fetching data from database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = hometmp.Execute(w, P)
	if err != nil {
		log.Fatal("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func insertPost(db *sql.DB, id int, post string) error {
	insertapost := `INSERT INTO STORE(ID, POSTTEXT) VALUES(?, ?)`
	statement, err := db.Prepare(insertapost)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer statement.Close()

	_, err = statement.Exec(id, post)
	if err != nil {
		return fmt.Errorf("error executing statement: %v", err)
	}

	return nil
}

func displaydata(db *sql.DB, P *[]Post) error {
	rows, err := db.Query("SELECT * FROM STORE")
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var posttext string
		err := rows.Scan(&id, &posttext)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		newPost := Post{
			ID:   id,
			Text: posttext,
		}
		*P = append(*P, newPost)
		fmt.Println(id, " ", posttext)
	}

	return nil
}
