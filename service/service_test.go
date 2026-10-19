package service

import (
	"errors"
	"testing"

	"github.com/Carter907/go-solve/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	testDB.Exec("PRAGMA foreign_keys = ON;")
	if err := db.AutoMigrate(testDB); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}
	db.SetConnection(testDB)
	return testDB
}

func TestSignUpUser_And_LoginUser_Errors(t *testing.T) {
	setupServiceTestDB(t)

	// 1. SignUp User Success
	user, err := SignUpUser("service_alice", "password123")
	if err != nil {
		t.Fatalf("SignUpUser failed: %v", err)
	}
	if user == nil || user.Username != "service_alice" {
		t.Fatalf("unexpected user: %+v", user)
	}

	// 2. SignUp User Duplicate -> ErrUsernameTaken
	_, dupErr := SignUpUser("service_alice", "anotherpass")
	if dupErr == nil {
		t.Fatalf("expected error on duplicate signup, got nil")
	}
	if !errors.Is(dupErr, ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", dupErr)
	}

	// 3. Login User Success
	loggedIn, err := LoginUser("service_alice", "password123")
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}
	if loggedIn.ID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, loggedIn.ID)
	}

	// 4. Login User Non-existent -> ErrUsernameNotFound
	_, notFoundErr := LoginUser("ghost_user", "password123")
	if notFoundErr == nil {
		t.Fatalf("expected error for non-existent user, got nil")
	}
	if !errors.Is(notFoundErr, ErrUsernameNotFound) {
		t.Fatalf("expected ErrUsernameNotFound, got %v", notFoundErr)
	}

	// 5. Login User Wrong Password -> ErrPasswordIncorrect
	_, wrongPassErr := LoginUser("service_alice", "wrongpass")
	if wrongPassErr == nil {
		t.Fatalf("expected error for wrong password, got nil")
	}
	if !errors.Is(wrongPassErr, ErrPasswordIncorrect) {
		t.Fatalf("expected ErrPasswordIncorrect, got %v", wrongPassErr)
	}
}

func TestService_SessionLifecycle(t *testing.T) {
	setupServiceTestDB(t)

	user, err := SignUpUser("session_alice", "secret")
	if err != nil {
		t.Fatalf("SignUpUser failed: %v", err)
	}

	session, err := CreateSession(user.ID)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if session == nil || session.Token == "" {
		t.Fatalf("invalid session returned: %+v", session)
	}

	fetchedUser := GetUserBySessionToken(session.Token)
	if fetchedUser == nil {
		t.Fatalf("expected user for active session token, got nil")
	}
	if fetchedUser.Username != "session_alice" {
		t.Fatalf("expected username 'session_alice', got %q", fetchedUser.Username)
	}

	// Empty token check
	if GetUserBySessionToken("") != nil {
		t.Fatalf("expected nil for empty token")
	}

	// Delete session
	if err := DeleteSession(session.Token); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}
	if GetUserBySessionToken(session.Token) != nil {
		t.Fatalf("expected nil user after deleting session")
	}
}

func TestService_TaskProgressAndCatalog(t *testing.T) {
	testDB := setupServiceTestDB(t)
	_ = db.SeedTasks(testDB)

	tasks := GetAllTasks()
	if len(tasks) == 0 {
		t.Fatalf("expected seeded tasks")
	}

	task, err := GetTaskByID(tasks[0].ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}
	if task.ID != tasks[0].ID {
		t.Fatalf("expected task ID %d, got %d", tasks[0].ID, task.ID)
	}

	// Non-existent task returns ErrNotFound
	_, errBad := GetTaskByID(99999)
	if errBad == nil || !errors.Is(errBad, db.ErrNotFound) {
		t.Fatalf("expected db.ErrNotFound, got %v", errBad)
	}

	user, _ := SignUpUser("progress_alice", "secret")

	// Progress initial check (not started)
	initProg := GetTaskProgress(user.ID, task.ID)
	if initProg != nil {
		t.Fatalf("expected nil progress initially, got %+v", initProg)
	}

	// Insert progress
	prog, err := InsertTaskProgress(user.ID, task.ID, "in progress")
	if err != nil {
		t.Fatalf("InsertTaskProgress failed: %v", err)
	}
	if prog.Progress != "in progress" {
		t.Fatalf("expected 'in progress', got %q", prog.Progress)
	}

	// Update progress
	if err := UpdateTaskProgress(prog.ID, "completed"); err != nil {
		t.Fatalf("UpdateTaskProgress failed: %v", err)
	}

	allProgress := GetAllTaskProgress(user.ID)
	if len(allProgress) != 1 || allProgress[0].Progress != "completed" {
		t.Fatalf("unexpected allProgress: %+v", allProgress)
	}
}
