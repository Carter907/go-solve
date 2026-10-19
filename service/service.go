package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Carter907/go-solve/db"
	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/security"
)

var (
	ErrUsernameNotFound  = errors.New("username not found")
	ErrPasswordIncorrect = errors.New("password is incorrect")
	ErrUsernameTaken     = errors.New("username is taken")
)

func LoginUser(username, password string) (*model.User, error) {
	userDB, err := db.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrUsernameNotFound
		}
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if !security.CheckPasswordHash(password, userDB.Password) {
		return nil, ErrPasswordIncorrect
	}
	return userDB, nil
}

func SignUpUser(username, password string) (*model.User, error) {
	userDB, err := db.InsertUser(username, password)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			return nil, ErrUsernameTaken
		}
		log.Printf("Error signing up user: %v\n", err)
		return nil, fmt.Errorf("failed to sign up: %w", err)
	}

	return userDB, nil
}

func GetAllTaskProgress(userID uint) []model.TaskProgress {
	taskProgresses, err := db.GetTaskProgressByUserID(userID)
	if err != nil {
		log.Printf("failed to get all task progress: %s\n", err.Error())
		return nil
	}

	return taskProgresses
}

func GetTaskProgress(userID uint, taskID uint) *model.TaskProgress {
	taskProgress, err := db.GetTaskProgressByUserIDAndTaskID(userID, taskID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			log.Printf("failed to get task progress: %s\n", err.Error())
		}
		return nil
	}

	return taskProgress
}

func GetAllTasks() []model.Task {
	tasks, err := db.GetAllTasks()
	if err != nil {
		log.Printf("failed to get all tasks: %s\n", err.Error())
		return nil
	}

	return tasks
}

func GetTaskByID(taskID uint) (*model.Task, error) {
	task, err := db.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func InsertTaskProgress(userID uint, taskID uint, progress string) (*model.TaskProgress, error) {
	taskProgress, err := db.InsertTaskProgress(userID, taskID, progress)
	if err != nil {
		log.Printf("failed to insert task progress: %s\n", err.Error())
		return nil, err
	}

	return taskProgress, nil
}

func UpdateTaskProgress(taskProgressID uint, progress string) error {
	err := db.UpdateTaskProgress(taskProgressID, progress)
	if err != nil {
		log.Printf("failed to update task progress: %s\n", err.Error())
		return err
	}

	return nil
}

const SessionDuration = 30 * 24 * time.Hour

func CreateSession(userID uint) (*model.Session, error) {
	return db.CreateSession(userID, SessionDuration)
}

func GetUserBySessionToken(token string) *model.User {
	if token == "" {
		return nil
	}
	session, err := db.GetSessionByToken(token)
	if err != nil || session == nil {
		return nil
	}
	return &session.User
}

func DeleteSession(token string) error {
	if token == "" {
		return nil
	}
	return db.DeleteSession(token)
}
