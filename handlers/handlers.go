package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/service"
)

const SessionCookieName = "session_token"

// GetUserFromRequest extracts and authenticates the user from the session cookie.
func GetUserFromRequest(r *http.Request) *model.User {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	return service.GetUserBySessionToken(cookie.Value)
}

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	user := GetUserFromRequest(r)
	if user == nil {
		user = &model.User{}
	}

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
	default:
		t, err := template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl = t
	}

	tasks := service.GetAllTasks()
	taskDataSlice := make([]TaskData, 0, len(tasks))

	for _, task := range tasks {
		var taskProgress = ""

		if len(user.Username) >= 1 && user.ID > 0 {
			taskProgress = "not started"
			taskProgresses := service.GetAllTaskProgress(user.ID)
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

func EditorHandler(w http.ResponseWriter, r *http.Request) {
	var task *model.Task

	// Check if taskID is provided (preferred)
	taskIDStr := r.FormValue("taskID")
	if taskIDStr != "" {
		id, err := strconv.Atoi(taskIDStr)
		if err == nil {
			t, err := service.GetTaskByID(uint(id))
			if err == nil {
				task = t
			}
		}
	}

	// Fallback to taskIndex for backwards compatibility
	if task == nil {
		taskIndexStr := r.FormValue("taskIndex")
		if taskIndexStr != "" {
			index, err := strconv.Atoi(taskIndexStr)
			if err == nil {
				tasks := service.GetAllTasks()
				if index >= 0 && index < len(tasks) {
					task = &tasks[index]
				}
			}
		}
	}

	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/editor.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "editor", task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RunCodeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	editorContent := r.FormValue("editorContent")
	taskIDStr := r.FormValue("taskID")

	var task *model.Task
	if taskIDStr != "" {
		id, err := strconv.Atoi(taskIDStr)
		if err == nil {
			t, err := service.GetTaskByID(uint(id))
			if err == nil {
				task = t
			}
		}
	}

	if task == nil {
		// Fallback to first available task if no taskID provided
		tasks := service.GetAllTasks()
		if len(tasks) > 0 {
			task = &tasks[0]
		} else {
			http.Error(w, "No tasks available to run", http.StatusBadRequest)
			return
		}
	}

	// Clone task for evaluation to ensure starter code in catalog is never mutated
	evalTask := *task
	evalTask.Code = editorContent

	taskResult := service.RunCode(&evalTask)

	user := GetUserFromRequest(r)
	if user != nil && user.ID > 0 {
		taskProgress := service.GetTaskProgress(user.ID, task.ID)
		if taskProgress == nil && taskResult.Passed {
			service.InsertTaskProgress(user.ID, task.ID, "completed")
		} else if taskResult.Passed {
			service.UpdateTaskProgress(taskProgress.ID, "completed")
		} else if taskProgress == nil {
			service.InsertTaskProgress(user.ID, task.ID, "in progress")
		} else {
			service.UpdateTaskProgress(taskProgress.ID, "in progress")
		}
	}

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
		log.Printf("Failed to parse signup template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "signup-form", args)
	if err != nil {
		log.Printf("Failed to execute signup template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

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
		case service.UsernameNotFound:
			args.UsernameNotFound = true
		}
	} else if userDB != nil {
		session, err := service.CreateSession(userDB.ID)
		if err != nil {
			log.Printf("Failed to create session: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    session.Token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  session.ExpiresAt,
		})
		w.Header().Set("HX-Refresh", "true")
	}

	tmpl, err := template.ParseFiles("templates/login-form.html")
	if err != nil {
		log.Printf("Failed to parse login template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "login-form", args)
	if err != nil {
		log.Printf("Failed to execute login template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = service.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
