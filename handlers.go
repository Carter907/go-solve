package main

import (
	"fmt"
	"github.com/Carter907/go-solve/model"
	"html/template"
	"net/http"
	"path"
	"strconv"
)

func TaskHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task, currTaskIndex *int) {

	index, err := strconv.Atoi(r.FormValue("taskIndex"))
	if err != nil {
		http.Error(w, "Failed to convert the task index to an int", http.StatusInternalServerError)
		return
	}

	*currTaskIndex = index

	EditorTmpl(w, &tasks[index])
}

func RunCodeHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task, taskIndex int) {

	editorContent := r.FormValue("editorContent")

	tasks[taskIndex].Code = editorContent
	taskResult := RunCode(&tasks[taskIndex]) // goes to TestSolution

	// call the Console handler after retrieving the correct task

	ConsoleOutput(w, &taskResult)
}

func ConsoleOutput(w http.ResponseWriter, c *model.TaskResult) {

	_, err := fmt.Fprintf(w, "tests: \n%v\n\n err: \n%v", c.Out, c.Err)
	if err != nil {
		http.Error(w, "Ran into unexpected error:", http.StatusInternalServerError)
		return
	}
}

func EditorTmpl(w http.ResponseWriter, t *model.Task) {

	fp := path.Join("templates", "editor.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Sending editor template with task: \n%v\n", *t)
	err = tmpl.Execute(w, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
