package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Carter907/go-solve/db"
	"github.com/Carter907/go-solve/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	// Ensure template directory can be found regardless of test working directory
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		_ = os.Chdir("..")
	}

	// Initialize test in-memory database
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	testDB.Exec("PRAGMA foreign_keys = ON;")
	_ = db.AutoMigrate(testDB)
	_ = db.SeedTasks(testDB)
	db.SetConnection(testDB)
}

func TestBaseHandler_Anonymous(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	BaseHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", res.Code, res.Body.String())
	}

	body := res.Body.String()
	if !strings.Contains(body, "Go-solve") {
		t.Fatalf("expected body to contain 'Go-solve', got: %s", body)
	}
	if strings.Contains(body, "Logged in as:") && strings.Contains(body, "alice") {
		t.Fatalf("expected anonymous user, but saw logged in username")
	}
}

func TestBaseHandler_Authenticated(t *testing.T) {
	user, err := db.InsertUser("alice_auth_test", "password123")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	session, err := service.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: session.Token,
	})

	res := httptest.NewRecorder()
	BaseHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", res.Code, res.Body.String())
	}

	body := res.Body.String()
	if !strings.Contains(body, "alice_auth_test") {
		t.Fatalf("expected body to contain authenticated username 'alice_auth_test', got: %s", body)
	}
}

func TestMultiUserSessionIsolation(t *testing.T) {
	userAlice, _ := db.InsertUser("user_alice", "pw1")
	userBob, _ := db.InsertUser("user_bob", "pw2")

	sessionAlice, _ := service.CreateSession(userAlice.ID)
	sessionBob, _ := service.CreateSession(userBob.ID)

	// Alice request
	reqAlice, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	reqAlice.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionAlice.Token})
	resAlice := httptest.NewRecorder()
	BaseHandler(resAlice, reqAlice)

	bodyAlice := resAlice.Body.String()
	if !strings.Contains(bodyAlice, "user_alice") || strings.Contains(bodyAlice, "user_bob") {
		t.Fatalf("Alice's response contains Bob's data or missing Alice: %s", bodyAlice)
	}

	// Bob request
	reqBob, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	reqBob.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionBob.Token})
	resBob := httptest.NewRecorder()
	BaseHandler(resBob, reqBob)

	bodyBob := resBob.Body.String()
	if !strings.Contains(bodyBob, "user_bob") || strings.Contains(bodyBob, "user_alice") {
		t.Fatalf("Bob's response contains Alice's data or missing Bob: %s", bodyBob)
	}
}

func TestEditorHandler_TaskID_And_TaskIndex(t *testing.T) {
	tasks := service.GetAllTasks()
	if len(tasks) == 0 {
		t.Fatalf("no tasks seeded")
	}
	firstTask := tasks[0]

	// 1. Test by taskID
	reqID, _ := http.NewRequest("GET", "http://localhost:8080/editor?taskID=1", nil)
	resID := httptest.NewRecorder()
	EditorHandler(resID, reqID)
	if resID.Code != http.StatusOK {
		t.Fatalf("expected status 200 for taskID, got %d", resID.Code)
	}

	// 2. Test by taskIndex
	reqIndex, _ := http.NewRequest("GET", "http://localhost:8080/editor?taskIndex=0", nil)
	resIndex := httptest.NewRecorder()
	EditorHandler(resIndex, reqIndex)
	if resIndex.Code != http.StatusOK {
		t.Fatalf("expected status 200 for taskIndex, got %d", resIndex.Code)
	}
	if !strings.Contains(resIndex.Body.String(), firstTask.Title) {
		t.Fatalf("expected editor body to contain %q, got: %s", firstTask.Title, resIndex.Body.String())
	}

	// 3. Test non-existent task
	reqBad, _ := http.NewRequest("GET", "http://localhost:8080/editor?taskID=99999", nil)
	resBad := httptest.NewRecorder()
	EditorHandler(resBad, reqBad)
	if resBad.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for invalid task, got %d", resBad.Code)
	}
}

func TestRunCodeHandler_Immutability(t *testing.T) {
	tasks := service.GetAllTasks()
	if len(tasks) == 0 {
		t.Fatalf("no tasks available")
	}
	originalCode := tasks[0].Code

	form := url.Values{}
	form.Set("taskID", "1")
	form.Set("editorContent", `package hello_world

func ReturnMessage() string {
	return "Hello World"
}
`)

	req, _ := http.NewRequest("POST", "http://localhost:8080/run-code", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res := httptest.NewRecorder()
	RunCodeHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", res.Code, res.Body.String())
	}

	if !strings.Contains(res.Body.String(), "PASS") {
		t.Fatalf("expected tests to pass, got: %s", res.Body.String())
	}

	// Verify starter code was NOT mutated in memory/database
	tasksAfter := service.GetAllTasks()
	if tasksAfter[0].Code != originalCode {
		t.Fatalf("starter code in catalog was mutated! Expected %q, got %q", originalCode, tasksAfter[0].Code)
	}
}

func TestLogoutHandler(t *testing.T) {
	user, _ := db.InsertUser("logout_user", "pw")
	session, _ := service.CreateSession(user.ID)

	req, _ := http.NewRequest("POST", "http://localhost:8080/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: session.Token,
	})

	res := httptest.NewRecorder()
	LogoutHandler(res, req)

	if res.Code != http.StatusSeeOther {
		t.Fatalf("expected HTTP 303 StatusSeeOther on logout, got %d", res.Code)
	}

	// Verify session token is removed from database
	if userAfter := service.GetUserBySessionToken(session.Token); userAfter != nil {
		t.Fatalf("expected session to be deleted from database after logout, but user still found")
	}

	// Verify expired cookie header is set
	cookies := res.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil || sessionCookie.MaxAge >= 0 {
		t.Fatalf("expected session_token cookie to be expired with negative MaxAge")
	}
}
