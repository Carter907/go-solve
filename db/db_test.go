package db

import (
	"testing"
	"time"

	"github.com/Carter907/go-solve/security"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	if err := AutoMigrate(testDB); err != nil {
		t.Fatalf("failed to auto migrate test db: %v", err)
	}

	SetConnection(testDB)
	return testDB
}

func TestAutoMigrateAndSeed(t *testing.T) {
	testDB := setupTestDB(t)

	// Clean tasks table for isolated seed test
	testDB.Exec("DELETE FROM task")

	if err := SeedTasks(testDB); err != nil {
		t.Fatalf("failed to seed tasks: %v", err)
	}

	tasks, err := GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != 7 {
		t.Fatalf("expected 7 tasks after seed, got %d", len(tasks))
	}

	// Idempotency check: running seed again should not duplicate tasks
	if err := SeedTasks(testDB); err != nil {
		t.Fatalf("second SeedTasks call failed: %v", err)
	}

	tasksAfter, err := GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasksAfter) != 7 {
		t.Fatalf("expected still 7 tasks after re-seed, got %d", len(tasksAfter))
	}
}

func TestUserCRUD(t *testing.T) {
	setupTestDB(t)

	username := "test_gopher"
	password := "supersecret123"

	// 1. Insert user
	user, err := InsertUser(username, password)
	if err != nil {
		t.Fatalf("InsertUser failed: %v", err)
	}

	if user.ID == 0 {
		t.Fatalf("expected user ID > 0, got %d", user.ID)
	}
	if user.Username != username {
		t.Fatalf("expected username %q, got %q", username, user.Username)
	}
	if !security.CheckPasswordHash(password, user.Password) {
		t.Fatalf("stored password hash does not verify against original password")
	}

	// 2. Query user by username
	fetchedUser, rowErr := GetUserByUsername(username)
	if rowErr != nil {
		t.Fatalf("GetUserByUsername failed: %v", rowErr)
	}
	if fetchedUser.ID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, fetchedUser.ID)
	}

	// 3. Duplicate user insertion should fail with RowNotUnique
	_, dupErr := InsertUser(username, "anotherpassword")
	if dupErr == nil {
		t.Fatalf("expected error on duplicate username, got nil")
	}
	rowError, ok := dupErr.(*RowError)
	if !ok || rowError.Status != RowNotUnique {
		t.Fatalf("expected RowError with RowNotUnique, got %v", dupErr)
	}

	// 4. Non-existent user should fail with RowNotFound
	_, notFoundErr := GetUserByUsername("non_existent_user_999")
	if notFoundErr == nil {
		t.Fatalf("expected error for non-existent user, got nil")
	}
	if notFoundErr.Status != RowNotFound {
		t.Fatalf("expected RowNotFound status, got %d", notFoundErr.Status)
	}
}

func TestTaskProgressCRUD(t *testing.T) {
	testDB := setupTestDB(t)

	// Seed user and task
	user, err := InsertUser("prog_user", "password123")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	_ = SeedTasks(testDB)
	tasks, _ := GetAllTasks()
	if len(tasks) == 0 {
		t.Fatalf("no tasks available for progress test")
	}
	taskID := tasks[0].ID

	// 1. Insert progress
	prog, err := InsertTaskProgress(user.ID, taskID, "in progress")
	if err != nil {
		t.Fatalf("InsertTaskProgress failed: %v", err)
	}
	if prog.Progress != "in progress" {
		t.Fatalf("expected progress 'in progress', got %q", prog.Progress)
	}

	// 2. Fetch by User ID and Task ID
	fetched, rowErr := GetTaskProgressByUserIDAndTaskID(user.ID, taskID)
	if rowErr != nil {
		t.Fatalf("GetTaskProgressByUserIDAndTaskID failed: %v", rowErr)
	}
	if fetched.ID != prog.ID {
		t.Fatalf("expected progress ID %d, got %d", prog.ID, fetched.ID)
	}

	// 3. Update progress
	if err := UpdateTaskProgress(prog.ID, "completed"); err != nil {
		t.Fatalf("UpdateTaskProgress failed: %v", err)
	}

	updated, rowErr := GetTaskProgressByUserIDAndTaskID(user.ID, taskID)
	if rowErr != nil {
		t.Fatalf("GetTaskProgressByUserIDAndTaskID after update failed: %v", rowErr)
	}
	if updated.Progress != "completed" {
		t.Fatalf("expected updated progress 'completed', got %q", updated.Progress)
	}

	// 4. Query all progress for user
	allProg, rowErr := GetTaskProgressByUserID(user.ID)
	if rowErr != nil {
		t.Fatalf("GetTaskProgressByUserID failed: %v", rowErr)
	}
	if len(allProg) != 1 {
		t.Fatalf("expected 1 progress record, got %d", len(allProg))
	}
}

func TestSessionCRUD(t *testing.T) {
	setupTestDB(t)

	user, err := InsertUser("session_user", "securepass")
	if err != nil {
		t.Fatalf("failed to insert user for session test: %v", err)
	}

	// 1. Create active session
	session, err := CreateSession(user.ID, 1*time.Hour)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if len(session.Token) != 64 {
		t.Fatalf("expected 64 hex characters token, got length %d", len(session.Token))
	}
	if session.UserID != user.ID {
		t.Fatalf("expected session UserID %d, got %d", user.ID, session.UserID)
	}

	// 2. Retrieve active session
	fetched, err := GetSessionByToken(session.Token)
	if err != nil {
		t.Fatalf("GetSessionByToken failed: %v", err)
	}
	if fetched.User.Username != "session_user" {
		t.Fatalf("expected preloaded user 'session_user', got %q", fetched.User.Username)
	}

	// 3. Create expired session and verify it is not retrieved
	expiredSession, err := CreateSession(user.ID, -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create expired session: %v", err)
	}
	_, err = GetSessionByToken(expiredSession.Token)
	if err == nil {
		t.Fatalf("expected GetSessionByToken to fail for expired session, but succeeded")
	}

	// 4. Delete session on logout
	if err := DeleteSession(session.Token); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}
	_, err = GetSessionByToken(session.Token)
	if err == nil {
		t.Fatalf("expected GetSessionByToken to fail after delete, but succeeded")
	}

	// 5. Delete expired sessions cleanup
	if err := DeleteExpiredSessions(); err != nil {
		t.Fatalf("DeleteExpiredSessions failed: %v", err)
	}
}

func TestGetTaskByID(t *testing.T) {
	testDB := setupTestDB(t)
	_ = SeedTasks(testDB)

	tasks, _ := GetAllTasks()
	if len(tasks) == 0 {
		t.Fatalf("expected seeded tasks")
	}

	task, rowErr := GetTaskByID(tasks[0].ID)
	if rowErr != nil {
		t.Fatalf("GetTaskByID failed: %v", rowErr)
	}
	if task.Title != tasks[0].Title {
		t.Fatalf("expected title %q, got %q", tasks[0].Title, task.Title)
	}

	_, rowErr = GetTaskByID(99999)
	if rowErr == nil || rowErr.Status != RowNotFound {
		t.Fatalf("expected RowNotFound for non-existent task ID")
	}
}
