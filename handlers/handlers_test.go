package handlers

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
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

func TestProblemHandler_Routing(t *testing.T) {
	tasks := service.GetAllTasks()
	if len(tasks) == 0 {
		t.Fatalf("no tasks seeded")
	}
	firstTask := tasks[0]

	// 1. Test by problem path
	reqID, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/problems/%d", firstTask.ID), nil)
	reqID.SetPathValue("id", strconv.Itoa(int(firstTask.ID)))
	resID := httptest.NewRecorder()
	ProblemHandler(resID, reqID)
	if resID.Code != http.StatusOK {
		t.Fatalf("expected status 200 for problem path, got %d", resID.Code)
	}
	if !strings.Contains(resID.Body.String(), firstTask.Title) {
		t.Fatalf("expected problem page to contain title %q, got: %s", firstTask.Title, resID.Body.String())
	}

	// 2. Test non-existent problem
	reqBad, _ := http.NewRequest("GET", "http://localhost:8080/problems/99999", nil)
	reqBad.SetPathValue("id", "99999")
	resBad := httptest.NewRecorder()
	ProblemHandler(resBad, reqBad)
	if resBad.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for invalid problem, got %d", resBad.Code)
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

func TestBaseHandler_CategoriesAndProgress(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	res := httptest.NewRecorder()
	BaseHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	body := html.UnescapeString(res.Body.String())
	categories := []string{"Basics", "Strings", "Slices & Sorting", "Maps & Structs"}
	for _, cat := range categories {
		if !strings.Contains(body, cat) {
			t.Errorf("expected body to contain category %q, got: %s", cat, body)
		}
	}

	// Verify toolbar controls
	if !strings.Contains(body, "problem-search") {
		t.Errorf("expected body to contain problem-search input")
	}
	if !strings.Contains(body, "difficulty-filters") {
		t.Errorf("expected body to contain difficulty-filters toolbar")
	}
}

func TestProblemHandler_BreadcrumbsAndNavigation(t *testing.T) {
	tasks := service.GetAllTasks()
	if len(tasks) < 2 {
		t.Fatalf("expected at least 2 tasks to test navigation, got %d", len(tasks))
	}

	// Request second task so it has both Prev and Next links
	secondTask := tasks[1]
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/problems/%d", secondTask.ID), nil)
	req.SetPathValue("id", strconv.Itoa(int(secondTask.ID)))
	res := httptest.NewRecorder()
	ProblemHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	body := res.Body.String()
	if !strings.Contains(body, "Back to Problems") {
		t.Errorf("expected editor body to contain 'Back to Problems' breadcrumb link")
	}
	if !strings.Contains(body, secondTask.Title) {
		t.Errorf("expected editor body to contain task title %q", secondTask.Title)
	}
	if !strings.Contains(body, secondTask.Category) {
		t.Errorf("expected editor body to contain task category %q", secondTask.Category)
	}
	if !strings.Contains(body, "Prev") {
		t.Errorf("expected editor body to contain 'Prev' button")
	}
	if !strings.Contains(body, "Next") {
		t.Errorf("expected editor body to contain 'Next' button")
	}
	// Verify self-hosted Ace assets are linked
	if !strings.Contains(body, "/static/ace/ace.js") {
		t.Errorf("expected problem page to link to local /static/ace/ace.js")
	}
	if !strings.Contains(body, "/static/ace/mode-golang.js") {
		t.Errorf("expected problem page to link to local /static/ace/mode-golang.js")
	}
}

