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
	"strconv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	tasks := LoadTasks()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/run-code", func(w http.ResponseWriter, r *http.Request) {
		editorContent := r.FormValue("editorContent")
		taskIndex, err := strconv.Atoi(r.FormValue("taskIndex"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		tasks[taskIndex].Code = editorContent
		fmt.Println(tasks[taskIndex].Code)
		out, errOut := editor.RunCode(&tasks[taskIndex])
		Console(w, r, &console.Console{
			Out: out.String(),
			Err: errOut.String(),
		})
	})

	http.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		index, err := strconv.Atoi(r.FormValue("taskIndex"))
		fmt.Println(index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		Editor(w, r, &tasks[index])
	})
	http.ListenAndServe(":"+port, nil)
}

func LoadTasks() (tasks []task.Task) {
	file, err := os.Open("./data/tasks.json")
	defer file.Close()
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tasks)
	if err != nil {
		log.Fatalf("failed to deserialize json: %v", err)
	}
	return
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
