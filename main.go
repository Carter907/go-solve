package main

import (
	"fmt"
	"github.com/Carter907/go-solve/model"
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

	tasks := GetTasks(GetConnection())

	user := &model.User{
		ID:       0,
		Username: "",
		Password: "",
	}
	currTaskIndex := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		BaseHandler(w, r, tasks, user)
	})

	http.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {

		TaskHandler(w, r, tasks, &currTaskIndex)
	})

	http.HandleFunc("/run-code", func(w http.ResponseWriter, r *http.Request) {

		RunCodeHandler(w, r, tasks, currTaskIndex)
	})

	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {

		SignupHandler(w, r)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		LoginHandler(w, r, user)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {

		LogoutHandler(w, r, user)
	})

	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Fatalln("Failed to run the server:", err)
		return
	}
}
