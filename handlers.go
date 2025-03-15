package main

import (
	"fmt"
	"github.com/Carter907/go-solve/model"
	"html/template"
	"net/http"
	"strconv"
)

func BaseHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task, user *model.User) {
	path := r.URL.Path[1:]
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
	type BaseArgs struct {
		Path  string
		Tasks []model.Task
		User  *model.User
	}
	var tmpl *template.Template

	switch path {
	case "signup-form":
		var t, err = template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl = t
		break
	default:
		t, err := template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl = t
		break
	}

	err := tmpl.Execute(w, BaseArgs{
		Path:  path,
		Tasks: tasks,
		User:  user,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	err := r.ParseForm()
	if err != nil {

		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	editorContent := r.FormValue("editorContent")

	tasks[taskIndex].Code = editorContent
	taskResult := RunCode(&tasks[taskIndex]) // goes to TestSolution

	_, err = fmt.Fprintf(w, "tests: \n%v\n\n err: \n%v", taskResult.Out, taskResult.Err)
	if err != nil {
		http.Error(w, "Ran into unexpected error:", http.StatusInternalServerError)
		return
	}
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm-password")
	if password != confirmPassword {
		http.Error(w, "Failed to signup, passwords don't match.", http.StatusBadRequest)
		return
	}
	if SignUpUser(GetConnection(), username, password) {

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	} else {

		http.Error(w, "Failed to Sign up.", http.StatusInternalServerError)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request, user *model.User) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println("logging in as", username, password)

	if userDB := GetUser(GetConnection(), username, password); userDB != nil {

		user.Username = userDB.Username
		user.Password = userDB.Password

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	} else {
		http.Error(w, "Failed to login, you are not authenticated.", http.StatusUnauthorized)
	}

}

func LogoutHandler(w http.ResponseWriter, r *http.Request, user *model.User) {

	fmt.Println("logging out")

	user.Username = ""
	user.Password = ""

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
