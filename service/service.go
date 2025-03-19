package service

import (
	"errors"
	"fmt"
	"github.com/Carter907/go-solve/db"
	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/security"
	"log"
)

func LoginUser(username, password string) (*model.User, *LoginError) {

	userDB, err := db.GetUserByUsername(db.GetConnection(), username)
	if err != nil {
		return nil, &LoginError{
			Status:  UsernameNotFound,
			Message: err.Message,
		}
	}

	if !security.CheckPasswordHash(password, userDB.Password) {
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

func SignUpUser(username, password string) (*model.User, *SignUpError) {
	userDB, err := db.InsertUser(db.GetConnection(), username, password)
	if err != nil {
		var rowError *db.RowError
		if errors.As(err, &rowError) && rowError.Status == db.RowNotUnique {
			return nil, &SignUpError{
				Status:  UsernameTaken,
				Message: "username is taken",
			}
		} else {

			log.Fatalln("yep it's here", err)
			return nil, nil
		}
	}

	return userDB, nil
}

const (
	UsernameTaken = 1
)

type SignUpStatus uint

type SignUpError struct {
	Status  SignUpStatus
	Message string
}

func (t SignUpError) Error() string {
	return t.Message
}
