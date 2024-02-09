package funcs

import (
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
}
