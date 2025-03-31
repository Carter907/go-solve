package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/security"
	_ "github.com/mattn/go-sqlite3"
)

var (
	once sync.Once
	Conn *sql.DB
)

func NewConnection() *sql.DB {
	once.Do(func() {
		db, err := sql.Open("sqlite3", "./data/db")
		if err != nil {
			log.Fatal("Failed to connect to sqlite")
		}
		Conn = db
	})

	return Conn
}

func GetUserByUsername(username string) (*model.User, *RowError) {

	var id uint
	var usernameDb string
	var passwordDb string
	err := Conn.QueryRow("select * from user where username = (?)",
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

	var id uint
	var usernameDB string
	var passwordDB string

	err := Conn.QueryRow("select (username) from user where username = (?)",
		username).Scan()
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

	res, err := Conn.Exec("insert into user(username, password) values((?), (?))", username,
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

	err = Conn.QueryRow("select * from user where id = (?)",
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

func GetAllTasks() ([]model.Task, *RowError) {
	rows, err := Conn.Query("select * from task")
	if err != nil {
		log.Fatalln("failed to query:", err)
		return nil, nil
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
			return nil, &RowError{
				Status:  RowColumnMisuse,
				Message: fmt.Sprintln("failed to scan rows:", err.Error()),
			}
		}

		task := model.Task{
			ID:          id,
			Title:       title,
			Description: description,
			Difficulty:  difficulty,
			Code:        code,
			Objective:   objective,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func GetAllTaskProgresses() []model.TaskProgress {

	rows, err := Conn.Query("select * from task_progress")
	if err != nil {
		log.Fatalln("Failed to query for tasks progresses:", err)
		return nil
	}

	taskProgresses := make([]model.TaskProgress, 0)

	for rows.Next() {
		var id uint
		var userId uint
		var taskId uint
		var progress string

		err = rows.Scan(&id, &userId, &taskId, &progress)
		if err != nil {
			log.Fatalln("Failed to scan tasks progresses:", err)
			return nil
		}

		taskProgress := model.TaskProgress{
			ID:       id,
			UserID:   userId,
			TaskID:   taskId,
			Progress: progress,
		}
		taskProgresses = append(taskProgresses, taskProgress)
	}

	return taskProgresses
}

func GetTaskProgressByUserID(userID uint) ([]model.TaskProgress, *RowError) {

	rows, err := Conn.Query("select * from task_progress where user_id = (?)",
		userID,
	)
	if err != nil {
		log.Fatalln("failed to query:", err)
		return nil, nil
	}
	taskProgress := make([]model.TaskProgress, 0)

	for rows.Next() {
		var id uint
		var userId uint
		var taskId uint
		var progress string

		err := rows.Scan(&id, &userId, &taskId, &progress)
		if err != nil {
			return nil, &RowError{
				Status:  RowColumnMisuse,
				Message: fmt.Sprintln("failed to scan rows:", err.Error()),
			}
		}

		taskProgress = append(taskProgress, model.TaskProgress{
			ID:       id,
			UserID:   userId,
			TaskID:   taskId,
			Progress: progress,
		})
	}

	return taskProgress, nil
}

func GetTaskProgressByUserIDAndTaskID(userID uint, taskID uint) (*model.TaskProgress, *RowError) {
	var id uint
	var userId uint
	var taskId uint
	var progress string

	err := Conn.QueryRow("select * from task_progress where user_id = (?) and task_id = (?)",
		userID,
		taskID,
	).Scan(&id, &userId, &taskId, &progress)

	if err != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("failed to find task progress for user", userID, "and task", taskID),
		}
	}

	taskProgress := &model.TaskProgress{
		ID:       id,
		UserID:   userId,
		TaskID:   taskId,
		Progress: progress,
	}

	return taskProgress, nil
}

func InsertTaskProgress(userID uint, taskID uint, progress string) (*model.TaskProgress, error) {

	res, err := Conn.Exec("insert into task_progress(user_id, task_id, progress) values((?), (?), (?))", userID, taskID, progress)
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

	var ID uint
	var userIdDB uint
	var taskIdDB uint
	var progressDB string

	err = Conn.QueryRow("select * from task_progress where id = (?)",
		insertId).Scan(&ID, &userIdDB, &taskIdDB, &progressDB)
	if err != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintln("row error:", err.Error()),
		}
	}

	return &model.TaskProgress{
		ID:       ID,
		UserID:   userIdDB,
		TaskID:   taskIdDB,
		Progress: progressDB,
	}, nil
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

	_, err := Conn.Exec("update task_progress set progress = (?) where id = (?)",
		progress, taskProgressID)
	if err != nil {
		return &UpdateError{
			Status:  UpdateErr,
			Message: fmt.Sprintln("insert error:", err.Error()),
		}
	}
	return nil
}
