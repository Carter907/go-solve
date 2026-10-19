package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Carter907/go-solve/db"
	"github.com/Carter907/go-solve/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("starting go-solve on http://localhost:%s\n", port)

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	database := db.NewConnection()
	if err := db.AutoMigrate(database); err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	if err := db.SeedTasks(database); err != nil {
		log.Fatalf("Failed to seed tasks: %v", err)
	}

	http.HandleFunc("/", handlers.BaseHandler)
	http.HandleFunc("/problems/{id}", handlers.ProblemHandler)
	http.HandleFunc("/run-code", handlers.RunCodeHandler)
	http.HandleFunc("/signup", handlers.SignupHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/healthz", handlers.HealthCheckHandler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalln("Failed to run the server:", err)
		return
	}
}
