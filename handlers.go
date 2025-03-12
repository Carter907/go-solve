package main

import (
	"fmt"
	"github.com/Carter907/go-solve/model"
	"html/template"
	"net/http"
	"strconv"
)

func BaseHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task) {
	path := r.URL.Path[1:]

	type BaseArgs struct {
		Path  string
		Tasks []model.Task
	}
	switch path {
	case "signup":
		tmpl, err := template.New("base").ParseGlob("templates/base.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, BaseArgs{
			Path:  path,
			Tasks: nil,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		break
	default:
		funcMap := template.FuncMap{
			"dict": func(values ...interface{}) (map[string]interface{}, error) {
				if len(values)%2 != 0 {
					return nil, fmt.Errorf("invalid dict call")
				}
				dict := make(map[string]interface{}, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil, fmt.Errorf("dict keys must be strings")
					}
					dict[key] = values[i+1]
				}
				return dict, nil
			},
		}
		tmpl, err := template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, BaseArgs{
			Path:  path,
			Tasks: tasks,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		break
	}

}

func TaskHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task, currTaskIndex *int) {

	index, err := strconv.Atoi(r.FormValue("taskIndex"))
	if err != nil {
		http.Error(w, "Failed to convert the task index to an int", http.StatusInternalServerError)
		return
	}

	*currTaskIndex = index

	tmpl, err := template.ParseFiles("templates/editor.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	task := &tasks[index]

	fmt.Printf("Sending editor template with task: \n%v\n", *task)
	err = tmpl.ExecuteTemplate(w, "editor", task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RunCodeHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task, taskIndex int) {

	editorContent := r.FormValue("editorContent")

	tasks[taskIndex].Code = editorContent
	taskResult := RunCode(&tasks[taskIndex]) // goes to TestSolution

	_, err := fmt.Fprintf(w, "tests: \n%v\n\n err: \n%v", taskResult.Out, taskResult.Err)
	if err != nil {
		http.Error(w, "Ran into unexpected error:", http.StatusInternalServerError)
		return
	}
}
