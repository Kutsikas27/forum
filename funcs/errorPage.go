package funcs

import (
	"log"
	"net/http"
	"text/template"
)

var errorPage = template.Must(template.New("errorPage.html").ParseFiles("frontend/templates/errorPage.html"))

func HandleErrorPage(w http.ResponseWriter, r *http.Request) {
	err := errorPage.Execute(w, nil)
	if err != nil {
		log.Fatal("Error executing error template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
