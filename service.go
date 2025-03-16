package main

import (
	"database/sql"
	"fmt"
	"github.com/Carter907/go-solve/model"
	"log"
)

func LoginUser(username, password string) (*model.User, *LoginError) {

	userDB, err := GetUserByUsername(GetConnection(), username)
	if err != nil {
		return nil, &LoginError{
			Status:  UsernameNotFound,
			Message: err.Message,
		}
	}

	if !CheckPasswordHash(password, userDB.Password) {
		return nil, &LoginError{
			Status:  PasswordIncorrect,
			Message: fmt.Sprintln("password is incorrect"),
		}
	}
	return userDB, nil
}

const (
	PasswordIncorrect = 1
	UsernameNotFound  = 2
)

type LoginStatus uint

type LoginError struct {
	Status  LoginStatus
	Message string
}

func (t LoginError) Error() string {
	return t.Message
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
