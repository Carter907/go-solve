package main

import (
	"encoding/json"
	"fmt"
	"github.com/Carter907/go-solve/model/console"
	"github.com/Carter907/go-solve/model/editor"
	"github.com/Carter907/go-solve/model/task"
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

	var tasks []task.Task
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tasks)
	if err != nil {
		log.Fatalf("failed to deserialize json: %v", err)
	}
	file.Close()
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {
		Editor(w, r, &tasks[0])

	})
	http.HandleFunc("/hello-world/console", func(w http.ResponseWriter, r *http.Request) {
		editorContent := r.FormValue("editorContent")
		tasks[0].Code = editorContent
		fmt.Println(tasks[0].Code)
		out, errOut := editor.RunCode(&tasks[0])
		Console(w, r, &console.Console{
			Out: out.String(),
			Err: errOut.String(),
		})
	})
	http.HandleFunc("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		Editor(w, r, &tasks[0])
	})
	http.ListenAndServe(":"+port, nil)
}

func Console(w http.ResponseWriter, r *http.Request, c *console.Console) {
	fp := path.Join("templates", "console.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Editor(w http.ResponseWriter, r *http.Request, t *task.Task) {
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
