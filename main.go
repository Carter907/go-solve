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

	fmt.Println("starting go-solve on http://localhost:8080")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	tasks := LoadTasks()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/run-code", func(w http.ResponseWriter, r *http.Request) {

		editorContent := r.FormValue("editorContent")
		taskIndexStr := r.FormValue("taskIndex")

		fmt.Println(editorContent, taskIndexStr)

		var taskIndex, err = strconv.Atoi(taskIndexStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		tasks[taskIndex].Code = editorContent
		out, errOut := editor.RunCode(&tasks[taskIndex]) // goes to TestSolution

		// call the Console handler after retrieving the correct task

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
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		return
	}
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

	_, err := fmt.Fprintf(w, "tests: \n%v\n\n err: \n%v", c.Out, c.Err)
	if err != nil {
		fmt.Println("Ran into unexpected error:", err)
		return
	}
}

func Editor(w http.ResponseWriter, r *http.Request, t *task.Task) {
	fp := path.Join("templates", "editor.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("executing editor with task: \n%v\n", *t)
	err = tmpl.Execute(w, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
