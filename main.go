package main

import (
	"fmt"
	"forum/funcs"
	"log"
	"net/http"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	http.HandleFunc("/", funcs.HomeHandler)
	http.HandleFunc("/login", funcs.LoginHandler)
	http.HandleFunc("/logout", funcs.LogOut)
	http.HandleFunc("/create-post", funcs.TopicHandler)
	http.HandleFunc("/comments/", funcs.ThreadHandler)

	addr := ":8080"
	fmt.Printf("Forum running at localhost%s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
