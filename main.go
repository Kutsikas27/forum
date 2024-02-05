package main

import (
	"fmt"
	"forum/funcs"
	"log"
	"net/http"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/static"))))
	http.HandleFunc("/", funcs.Homepage)
	http.HandleFunc("/signin", funcs.LoginPage)

	addr := ":8080"
	fmt.Printf("Forum running at localhost%s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
