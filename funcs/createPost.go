package funcs

import (
	"database/sql"
	"log"
	"net/http"
	"text/template"

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
		r.ParseForm()
		title := r.Form.Get("title")
		text := r.Form.Get("text")
		category := r.Form.Get("category")

		insertPostIntoDB(title, text, category)
		http.Redirect(w, r, "/", http.StatusFound)
	}

}
func insertPostIntoDB(title, text, category string) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO post(title, text, category) VALUES (?, ?, ?)", title, text, category)
	if err != nil {
		log.Fatal(err)
	}
}
