package main

import (
	"encoding/json"
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

	tasks := GetTasksArray("./data/tasks.json")
	currTaskIndex := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		BaseHandler(w, r, tasks)
	})

	http.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {

		TaskHandler(w, r, tasks, &currTaskIndex)
	})

	http.HandleFunc("/run-code", func(w http.ResponseWriter, r *http.Request) {

		RunCodeHandler(w, r, tasks, currTaskIndex)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		_, err := fmt.Fprintln(w, "Welcome to login")
		if err != nil {
			log.Fatalln("Failed to reach the login page", http.StatusInternalServerError)
			return
		}
	})

	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Fatalln("Failed to run the server:", err)
		return
	}
}

func GetTasksArray(tasksPath string) (tasks []model.Task) {
	file, err := os.Open(tasksPath)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln("Failed to close path:", tasksPath)
		}
	}(file)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tasks)
	if err != nil {
		log.Fatalln("Failed to deserialize json:", err)
	}
	return
}
