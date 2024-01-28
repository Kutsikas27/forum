package main

import (
	"fmt"
	"forum/funcs"
	"log"
	"net/http"
)

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("/frontend/assets"))))
	http.HandleFunc("/", funcs.Homepage)
	addr := ":8080"
	fmt.Printf("Forum running at localhost%s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
