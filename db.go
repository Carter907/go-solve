package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func GetConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./data/users")
	if err != nil {
		log.Fatal("Failed to connect to sqlite")
		return nil
	}
	return db
}

type UserDB struct {
	ID       uint
	Username string
	Password string
}

func GetUser(db *sql.DB, username string, password string) (user *UserDB) {

	rows, err := db.Query("select * from user where username = (?) and password = (?)",
		username,
		password,
	)
	if err != nil {
		log.Fatalln("Failed to query the user with username and password.")
		return nil
	}

	for rows.Next() {
		var id uint
		var username string
		var password string
		err = rows.Scan(&username, &password, &id)
		if err != nil {
			log.Fatalln(err)
			return nil
		}

		user = &UserDB{
			ID:       id,
			Username: username,
			Password: password,
		}
	}
	return
}
