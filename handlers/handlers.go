package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Carter907/go-solve/db"
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

var funcMap = template.FuncMap{
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

type CategoryGroup struct {
	Name             string
	Description      string
	TotalCount       int
	CompletedCount   int
	PercentCompleted int
	Tasks            []TaskData
}

type EditorData struct {
	model.Task
	PrevTaskID uint
	NextTaskID uint
}

type BaseArgs struct {
	Path             string
	Categories       []CategoryGroup
	TaskData         []TaskData
	EditorData       EditorData
	TotalTasks       int
	CompletedTasks   int
	PercentCompleted int
	User             *model.User
}

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	user := GetUserFromRequest(r)
	if user == nil {
		user = &model.User{}
	}
	var tmpl *template.Template

	switch path {
	case "signup-form":
		t, err := template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
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

	progressMap := make(map[uint]string)
	if len(user.Username) >= 1 && user.ID > 0 {
		taskProgresses := service.GetAllTaskProgress(user.ID)
		for _, prog := range taskProgresses {
			progressMap[prog.TaskID] = prog.Progress
		}
	}

	categoryDescriptions := map[string]string{
		"Basics":           "Core syntax, basic types, return values, and arithmetic.",
		"Strings":          "UTF-8 manipulation, string traversal, and runes.",
		"Slices & Sorting": "Dynamic slices, sorting algorithms, and element searches.",
		"Maps & Structs":   "Key-value collections, structures, and frequency maps.",
	}
	categoryOrder := []string{"Basics", "Strings", "Slices & Sorting", "Maps & Structs"}

	categoryMap := make(map[string][]TaskData)
	totalCompleted := 0

	for _, task := range tasks {
		taskProgress := ""
		if len(user.Username) >= 1 && user.ID > 0 {
			taskProgress = "not started"
			if p, ok := progressMap[task.ID]; ok {
				taskProgress = p
			}
		}

		if taskProgress == "completed" {
			totalCompleted++
		}

		cat := task.Category
		if cat == "" {
			cat = "Basics"
		}

		td := TaskData{
			Task:         task,
			TaskProgress: taskProgress,
		}
		taskDataSlice = append(taskDataSlice, td)
		categoryMap[cat] = append(categoryMap[cat], td)
	}

	var categories []CategoryGroup
	seen := make(map[string]bool)

	for _, name := range categoryOrder {
		if tasksInCat, ok := categoryMap[name]; ok {
			completedInCat := 0
			for _, td := range tasksInCat {
				if td.TaskProgress == "completed" {
					completedInCat++
				}
			}
			percent := 0
			if len(tasksInCat) > 0 {
				percent = (completedInCat * 100) / len(tasksInCat)
			}
			desc := categoryDescriptions[name]
			if desc == "" {
				desc = "Practice exercises for " + name
			}
			categories = append(categories, CategoryGroup{
				Name:             name,
				Description:      desc,
				TotalCount:       len(tasksInCat),
				CompletedCount:   completedInCat,
				PercentCompleted: percent,
				Tasks:            tasksInCat,
			})
			seen[name] = true
		}
	}

	for name, tasksInCat := range categoryMap {
		if !seen[name] {
			completedInCat := 0
			for _, td := range tasksInCat {
				if td.TaskProgress == "completed" {
					completedInCat++
				}
			}
			percent := 0
			if len(tasksInCat) > 0 {
				percent = (completedInCat * 100) / len(tasksInCat)
			}
			categories = append(categories, CategoryGroup{
				Name:             name,
				Description:      "Practice exercises for " + name,
				TotalCount:       len(tasksInCat),
				CompletedCount:   completedInCat,
				PercentCompleted: percent,
				Tasks:            tasksInCat,
			})
		}
	}

	overallPercent := 0
	if len(tasks) > 0 {
		overallPercent = (totalCompleted * 100) / len(tasks)
	}

	err := tmpl.Execute(w, BaseArgs{
		Path:             path,
		Categories:       categories,
		TaskData:         taskDataSlice,
		TotalTasks:       len(tasks),
		CompletedTasks:   totalCompleted,
		PercentCompleted: overallPercent,
		User:             user,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ProblemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	if idStr == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 && parts[0] == "problems" {
			idStr = parts[1]
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	task, err := service.GetTaskByID(uint(id))
	if err != nil || task == nil {
		http.NotFound(w, r)
		return
	}

	allTasks := service.GetAllTasks()
	var prevTaskID, nextTaskID uint
	for i, t := range allTasks {
		if t.ID == task.ID {
			if i > 0 {
				prevTaskID = allTasks[i-1].ID
			}
			if i < len(allTasks)-1 {
				nextTaskID = allTasks[i+1].ID
			}
			break
		}
	}

	user := GetUserFromRequest(r)
	if user == nil {
		user = &model.User{}
	}

	editorData := EditorData{
		Task:       *task,
		PrevTaskID: prevTaskID,
		NextTaskID: nextTaskID,
	}

	tmpl, err := template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, BaseArgs{
		Path:       "editor",
		EditorData: editorData,
		User:       user,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sqlDB, err := db.Conn.DB()
	if err != nil || sqlDB.Ping() != nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:   "unhealthy",
			Database: "unreachable",
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:   "ok",
		Database: "connected",
	})
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
	if serr != nil && errors.Is(serr, service.ErrUsernameTaken) {
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
		if errors.Is(err1, service.ErrPasswordIncorrect) {
			args.PasswordIncorrect = true
		} else if errors.Is(err1, service.ErrUsernameNotFound) {
			args.UsernameNotFound = true
		} else {
			log.Printf("Failed to login user: %v\n", err1)
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
		w.Header().Set("HX-Redirect", "/")
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
