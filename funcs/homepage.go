package funcs

import (
	"net/http"
	"text/template"
)

var hometmp = template.Must(template.ParseGlob("frontend/templates/*.html"))

func Homepage(w http.ResponseWriter, r *http.Request) {
	hometmp.ExecuteTemplate(w, "index.html", nil)
}
