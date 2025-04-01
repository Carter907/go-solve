package db

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"sync"

	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/security"
	_ "github.com/mattn/go-sqlite3"
)

var (
	once sync.Once
	Conn *gorm.DB
)

func NewConnection() *gorm.DB {
	once.Do(func() {
		db, err := gorm.Open(sqlite.Open("./data/db"), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect to sqlite")
		}
		Conn = db
	})

	return Conn
}

func GetUserByUsername(username string) (*model.User, *RowError) {

	var user model.User

	err := Conn.Table("user").Where("username = ?", username).First(&user).Error	

	if err != nil {

		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("failed to find user with username", username),
		}
	}

	return &user, nil
}

const (
	RowNotFound     = 1
	RowNotUnique    = 2
	RowColumnMisuse = 3
)

type RowStatus uint

type RowError struct {
	Status  RowStatus
	Message string
}

func (r RowError) Error() string {
	return r.Message
}

func InsertUser(username string, password string) (*model.User, error) {

	var user model.User

	err := Conn.Table("user").Where("username = ?", username).First(&user).Error

	if err == nil {
		return nil, &RowError{
			Status:  RowNotUnique,
			Message: fmt.Sprintln("row error: user found"),
		}
	}
	password, err = security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	err = Conn.Table("user").Create(&model.User{Username: username, Password: password}).Error
	if err != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintln("insert error:", err.Error()),
		}
	}
	var newUser model.User
	result := Conn.Table("user").Last(&newUser)
	if result.Error != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintln("insert error:", result.Error),
		}
	}

	result = Conn.Table("user").Where("id = ?", newUser.ID).First(&user)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("row error:", result.Error),
		}
	}

	return &model.User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
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

func GetAllTasks() ([]model.Task, *RowError) {
	var tasks []model.Task
	result := Conn.Table("task").Find(&tasks)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowColumnMisuse,
			Message: fmt.Sprintln("failed to find tasks:", result.Error),
		}
	}

	return tasks, nil
}

func GetAllTaskProgresses() []model.TaskProgress {
	var taskProgresses []model.TaskProgress
	result := Conn.Table("task_progress").Find(&taskProgresses)
	if result.Error != nil {
		log.Fatalln("Failed to query task progresses:", result.Error)
		return nil
	}

	return taskProgresses
}

func GetTaskProgressByUserID(userID uint) ([]model.TaskProgress, *RowError) {
	var taskProgress []model.TaskProgress
	result := Conn.Table("task_progress").Where("user_id = ?", userID).Find(&taskProgress)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowColumnMisuse,
			Message: fmt.Sprintln("failed to find task progress:", result.Error),
		}
	}

	return taskProgress, nil
}

func GetTaskProgressByUserIDAndTaskID(userID uint, taskID uint) (*model.TaskProgress, *RowError) {
	var taskProgress model.TaskProgress
	result := Conn.Table("task_progress").Where("user_id = ? AND task_id = ?", userID, taskID).First(&taskProgress)

	if result.Error != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("failed to find task progress for user", userID, "and task", taskID),
		}
	}

	return &taskProgress, nil
}

func InsertTaskProgress(userID uint, taskID uint, progress string) (*model.TaskProgress, error) {

	taskProgress := model.TaskProgress{
		UserID:   userID,
		TaskID:   taskID,
		Progress: progress,
	}

	result := Conn.Table("task_progress").Create(&taskProgress)
	if result.Error != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintln("insert error:", result.Error),
		}
	}

	var createdTaskProgress model.TaskProgress
	result = Conn.Table("task_progress").Where("id = ?", taskProgress.ID).First(&createdTaskProgress)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("row error:", result.Error),
		}
	}

	return &createdTaskProgress, nil
}

const (
	UpdateErr = 1
)

type UpdateStatus uint

type UpdateError struct {
	Status  UpdateStatus
	Message string
}

func (r UpdateError) Error() string {
	return r.Message
}

func UpdateTaskProgress(taskProgressID uint, progress string) error {

	result := Conn.Table("task_progress").Where("id = ?", taskProgressID).Update("progress", progress)
	if result.Error != nil {
		return &UpdateError{
			Status:  UpdateErr,
			Message: fmt.Sprintln("insert error:", result.Error),
		}
	}
	return nil
}
