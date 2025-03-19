package db

import (
	"database/sql"
	"fmt"
	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/security"
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

func GetUserByUsername(db *sql.DB, username string) (*model.User, *RowError) {

	var id uint
	var usernameDb string
	var passwordDb string
	err := db.QueryRow("select * from user where username = (?)",
		username,
	).Scan(&usernameDb, &passwordDb, &id)
	if err != nil {

		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("failed to find user with username", username),
		}
	}

	return &model.User{
		ID:       id,
		Username: usernameDb,
		Password: passwordDb,
	}, nil
}

const (
	RowNotFound  = 1
	RowNotUnique = 2
)

type RowStatus uint

type RowError struct {
	Status  RowStatus
	Message string
}

func (r RowError) Error() string {
	return r.Message
}

func InsertUser(db *sql.DB, username string, password string) (*model.User, error) {

	var id uint
	var usernameDB string
	var passwordDB string

	err := db.QueryRow("select (username) from user where username = (?)",
		username).Scan()
	if err == nil {
		return nil, &RowError{
			Status:  RowNotUnique,
			Message: fmt.Sprintln("row error: ", err.Error()),
		}
	}
	password, err = security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	res, err := db.Exec("insert into user(username, password) values((?), (?))", username,
		password)
	if err != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintln("insert error:", err.Error()),
		}
	}
	insertId, err := res.LastInsertId()
	if err != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintln("insert error:", err.Error()),
		}
	}

	err = db.QueryRow("select * from user where id = (?)",
		insertId).Scan(&usernameDB, &passwordDB, &id)
	if err != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("row error:", err.Error()),
		}
	}

	return &model.User{
		ID:       id,
		Username: usernameDB,
		Password: passwordDB,
	}, nil
}

const (
	InsertErr = 1
)

type InsertStatus uint

type InsertError struct {
	Status  InsertStatus
	Message string
}

func (r InsertError) Error() string {
	return r.Message
}

func GetAllTasks(db *sql.DB) []model.Task {
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
