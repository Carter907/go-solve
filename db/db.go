package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Carter907/go-solve/model"
	"github.com/Carter907/go-solve/security"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	once sync.Once
	Conn *gorm.DB
)

// InitDB initializes a GORM SQLite connection with the provided database path.
func InitDB(dbPath string) (*gorm.DB, error) {
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	// Enable foreign key enforcement
	db.Exec("PRAGMA foreign_keys = ON;")

	return db, nil
}

// SetConnection explicitly sets the active package-level database connection.
func SetConnection(db *gorm.DB) {
	Conn = db
}

// NewConnection initializes or returns the singleton database connection.
func NewConnection() *gorm.DB {
	once.Do(func() {
		dbPath := os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			dbPath = "./data/go-solve.db"
		}

		db, err := InitDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		Conn = db
	})

	return Conn
}

// AutoMigrate migrates the database schema for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Task{},
		&model.TaskProgress{},
		&model.Session{},
	)
}

func GetUserByUsername(username string) (*model.User, *RowError) {
	var user model.User
	err := Conn.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintf("failed to find user with username %s", username),
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
	var existing model.User
	err := Conn.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil, &RowError{
			Status:  RowNotUnique,
			Message: fmt.Sprintf("row error: user %s already exists", username),
		}
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: username,
		Password: hashedPassword,
	}

	if err := Conn.Create(&user).Error; err != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintf("insert error: %v", err),
		}
	}

	return &user, nil
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
	result := Conn.Find(&tasks)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowColumnMisuse,
			Message: fmt.Sprintf("failed to find tasks: %v", result.Error),
		}
	}

	return tasks, nil
}

func GetTaskByID(taskID uint) (*model.Task, *RowError) {
	var task model.Task
	result := Conn.Where("id = ?", taskID).First(&task)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintf("failed to find task with ID %d: %v", taskID, result.Error),
		}
	}

	return &task, nil
}

func GetAllTaskProgresses() []model.TaskProgress {
	var taskProgresses []model.TaskProgress
	result := Conn.Find(&taskProgresses)
	if result.Error != nil {
		log.Printf("Failed to query task progresses: %v", result.Error)
		return nil
	}

	return taskProgresses
}

func GetTaskProgressByUserID(userID uint) ([]model.TaskProgress, *RowError) {
	var taskProgress []model.TaskProgress
	result := Conn.Where("user_id = ?", userID).Find(&taskProgress)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowColumnMisuse,
			Message: fmt.Sprintf("failed to find task progress: %v", result.Error),
		}
	}

	return taskProgress, nil
}

func GetTaskProgressByUserIDAndTaskID(userID uint, taskID uint) (*model.TaskProgress, *RowError) {
	var taskProgress model.TaskProgress
	result := Conn.Where("user_id = ? AND task_id = ?", userID, taskID).First(&taskProgress)
	if result.Error != nil {
		return nil, &RowError{
			Status:  RowNotFound,
			Message: fmt.Sprintf("failed to find task progress for user %d and task %d", userID, taskID),
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

	result := Conn.Create(&taskProgress)
	if result.Error != nil {
		return nil, &InsertError{
			Status:  InsertErr,
			Message: fmt.Sprintf("insert error: %v", result.Error),
		}
	}

	return &taskProgress, nil
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
	result := Conn.Model(&model.TaskProgress{}).Where("id = ?", taskProgressID).Update("progress", progress)
	if result.Error != nil {
		return &UpdateError{
			Status:  UpdateErr,
			Message: fmt.Sprintf("update error: %v", result.Error),
		}
	}
	return nil
}

// CreateSession generates a secure random token and creates a new session for the given user.
func CreateSession(userID uint, duration time.Duration) (*model.Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	session := model.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(duration),
	}

	if err := Conn.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSessionByToken retrieves an active (unexpired) session by token, preloading the associated User.
func GetSessionByToken(token string) (*model.Session, error) {
	var session model.Session
	err := Conn.Preload("User").Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession removes a session by token.
func DeleteSession(token string) error {
	return Conn.Where("token = ?", token).Delete(&model.Session{}).Error
}

// DeleteExpiredSessions cleans up expired sessions from the database.
func DeleteExpiredSessions() error {
	return Conn.Where("expires_at <= ?", time.Now()).Delete(&model.Session{}).Error
}
