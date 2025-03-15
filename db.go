package main

import (
	"database/sql"
	"github.com/Carter907/go-solve/model"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func GetConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./data/db")
	if err != nil {
		log.Fatal("Failed to connect to sqlite")
		return nil
	}
	return db
}

func GetUser(db *sql.DB, username string, password string) *model.User {

	rows, err := db.Query("select * from user where username = (?)",
		username,
	)
	if err != nil {
		log.Fatalln("Failed to query the user with username and password.")
		return nil
	}

	for rows.Next() {
		var id uint
		var username string
		var hashedPassword string
		err = rows.Scan(&username, &hashedPassword, &id)
		if err != nil {
			log.Fatalln(err)
			return nil
		}
		if !CheckPasswordHash(password, hashedPassword) {
			log.Fatalln("Failed to log in. Incorrect password")
			return nil
		}

		return &model.User{
			ID:       id,
			Username: username,
			Password: password,
		}
	}
	return nil
}

func SignUpUser(db *sql.DB, username string, password string) bool {

	rows, err := db.Query("select (username) from user where username = (?)",
		username)
	if err != nil {
		log.Fatalln("Failed to query for usernames", err)
	}

	for rows.Next() {
		var username string

		err := rows.Scan(&username)
		if err != nil {
			log.Fatalln("Failed to scan username for query:", err)
			return false
		}
		log.Fatalln("Username", username, "is taken.")
		return false
	}

	// hash password

	password, err = HashPassword(password)
	if err != nil {
		log.Fatalln("Failed to hash password:", err)
		return false
	}

	_, err = db.Exec("insert into user(username, password) values((?), (?))", username, password)
	if err != nil {
		log.Fatalln("Failed to insert new user", err)
		return false
	}

	return true
}

func GetTasks(db *sql.DB) []model.Task {
	rows, err := db.Query("select * from task")
	if err != nil {
		log.Fatalln("Failed to query for tasks:", err)
		return nil
	}

	tasks := make([]model.Task, 0)

	for rows.Next() {
		var id uint
		var title string
		var description string
		var difficulty string
		var code string
		var objective string

		err = rows.Scan(&id, &title, &description, &difficulty, &code, &objective)
		if err != nil {
			log.Fatalln("Failed to scan tasks:", err)
			return nil
		}

		task := model.Task{
			Title:       title,
			Description: description,
			Difficulty:  difficulty,
			Code:        code,
			Objective:   objective,
		}
		tasks = append(tasks, task)
	}

	return tasks
}
