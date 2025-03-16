package main

import (
	"database/sql"
	"fmt"
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
	RowNotFound = 1
)

type RowStatus uint

type RowError struct {
	Status  RowStatus
	Message string
}

func (r RowError) Error() string {
	return r.Message
}
