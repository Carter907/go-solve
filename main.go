package main

import (
	"fmt"
	"github.com/Carter907/go-solve/db"
	"github.com/Carter907/go-solve/handlers"
	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/service"
	"log"
	"net/http"
	"os"
)

func main() {

	fmt.Println("starting go-solve on http://localhost:8080")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	db.NewConnection()
	tasks := service.GetAllTasks()

	user := &model.User{
		ID:       0,
		Username: "",
		Password: "",
	}
	currTaskIndex := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.BaseHandler(w, r, tasks, user)
	})

	http.HandleFunc("/editor", func(w http.ResponseWriter, r *http.Request) {

		handlers.EditorHandler(w, r, tasks, &currTaskIndex)
	})

	http.HandleFunc("/run-code", func(w http.ResponseWriter, r *http.Request) {

		handlers.RunCodeHandler(w, r, &tasks[currTaskIndex], user)
	})

	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {

		handlers.SignupHandler(w, r)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		handlers.LoginHandler(w, r, user)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {

		handlers.LogoutHandler(w, r, user)
	})

	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Fatalln("Failed to run the server:", err)
		return
	}
}
