package handlers

import (
	"fmt"
	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/service"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func BaseHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task,
	user *model.User) {
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
	type TaskData struct {
		Task         model.Task
		TaskProgress string
	}
	type BaseArgs struct {
		Path     string
		TaskData []TaskData
		User     *model.User
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

	taskDataSlice := make([]TaskData, 0)

	for _, task := range tasks {
		var taskProgress = ""

		fmt.Println("task: ", task.ID)

		if len(user.Username) >= 1 {
			taskProgress = "not started"

			taskProgresses := service.GetUserProgress(user.ID)

			for _, taskProg := range taskProgresses {
				if task.ID == taskProg.TaskID {
					taskProgress = taskProg.Progress
					break
				}
			}
		}

		taskData := TaskData{
			Task:         task,
			TaskProgress: taskProgress,
		}

		taskDataSlice = append(taskDataSlice, taskData)
	}

	err := tmpl.Execute(w, BaseArgs{
		Path:     path,
		TaskData: taskDataSlice,
		User:     user,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func EditorHandler(w http.ResponseWriter, r *http.Request, tasks []model.Task, currTaskIndex *int) {

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

	fmt.Printf("Sending editor template with task: \n%s\n", task.Code)
	err = tmpl.ExecuteTemplate(w, "editor", task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RunCodeHandler(w http.ResponseWriter, r *http.Request, task *model.Task) {

	err := r.ParseForm()
	if err != nil {

		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	editorContent := r.FormValue("editorContent")

	task.Code = editorContent
	taskResult := service.RunCode(task) // goes to TestSolution

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
	fmt.Println("signing up as ", username, confirmPassword)
	type SignUpArgs struct {
		UsernameTaken      bool
		PasswordsDontMatch bool
		SignUpSuccess      bool
	}

	args := SignUpArgs{
		UsernameTaken:      false,
		PasswordsDontMatch: password != confirmPassword,
		SignUpSuccess:      false,
	}

	_, serr := service.SignUpUser(username, password)
	if serr != nil && serr.Status == service.UsernameTaken {
		args.UsernameTaken = true
	}
	if !(args.UsernameTaken || args.PasswordsDontMatch) {
		args.SignUpSuccess = true
	}

	tmpl, err := template.ParseFiles("templates/signup-form.html")
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "signup-form", args)
	if err != nil {
		log.Fatalln(err)
		return
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

	type LoginArgs struct {
		UsernameNotFound  bool
		PasswordIncorrect bool
	}
	args := LoginArgs{
		UsernameNotFound:  false,
		PasswordIncorrect: false,
	}
	userDB, err1 := service.LoginUser(username, password)
	if err1 != nil {

		switch err1.Status {
		case service.PasswordIncorrect:
			args.PasswordIncorrect = true
			break
		case service.UsernameNotFound:
			args.UsernameNotFound = true
			break
		}
	} else {

		w.Header().Set("HX-Refresh", "true")
	}

	if userDB != nil {

		user.ID = userDB.ID
		user.Username = userDB.Username
		user.Password = userDB.Password

	}
	tmpl, err := template.ParseFiles("templates/login-form.html")
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "login-form", args)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, user *model.User) {

	fmt.Println("logging out")

	user.Username = ""
	user.Password = ""

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
