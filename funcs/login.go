package funcs

import (
	"log"
	"net/http"
	"text/template"
)

var logintmp = template.Must(template.New("login.html").ParseFiles("frontend/templates/login.html"))

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/login" {
		if r.Method == "GET" {
			err := logintmp.Execute(w, nil)
			if err != nil {
				log.Fatal("Error executing template:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
	}
}
