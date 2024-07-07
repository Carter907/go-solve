package main

import (
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/editor.html")
	})
	http.ListenAndServe(":"+port, nil)
}
