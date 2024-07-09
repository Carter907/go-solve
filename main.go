package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	file, err := os.Open("./data/tasks.json")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
		return
	}

	var tasks []Task
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tasks)
	if err != nil {
		log.Fatalf("failed to deserialize json: %v", err)
	}
	file.Close()
	fmt.Println(&tasks)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {
		Editor(w, r, &tasks[0])

	})
	http.HandleFunc("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		Editor(w, r, &tasks[0])
	})
	http.ListenAndServe(":"+port, nil)
}

type Task struct {
	Title string `json:"title"`
	Task  string `json:"task"`
	Code  string `json:"code"`
}

func Editor(w http.ResponseWriter, r *http.Request, t *Task) {
	fp := path.Join("templates", "editor.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
