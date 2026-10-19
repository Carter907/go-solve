package handlers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Carter907/go-solve/model"
)

func TestTemplateRegistry_RenderHome(t *testing.T) {
	registry, err := NewTemplateRegistry("")
	if err != nil {
		t.Fatalf("failed to create template registry: %v", err)
	}

	// 1. Anonymous user
	var bufAnonymous bytes.Buffer
	err = registry.RenderHome(&bufAnonymous, HomeViewData{
		User:             &model.User{},
		Categories:       nil,
		TotalTasks:       5,
		CompletedTasks:   0,
		PercentCompleted: 0,
	})
	if err != nil {
		t.Fatalf("RenderHome (anonymous) failed: %v", err)
	}
	outputAnonymous := bufAnonymous.String()
	if !strings.Contains(outputAnonymous, "Go-Solve") {
		t.Errorf("expected home page to contain 'Go-Solve'")
	}
	if !strings.Contains(outputAnonymous, "Sign in") {
		t.Errorf("expected home page to contain 'Sign in' button for anonymous user")
	}

	// 2. Authenticated user
	var bufAuth bytes.Buffer
	authCategory := CategoryGroup{
		Name:             "Basics",
		Description:      "Basic concepts",
		TotalCount:       1,
		CompletedCount:   1,
		PercentCompleted: 100,
		Tasks: []TaskData{
			{
				Task: model.Task{
					ID:         1,
					Title:      "Hello World",
					Difficulty: "easy",
					Category:   "Basics",
				},
				TaskProgress: "completed",
			},
		},
	}

	err = registry.RenderHome(&bufAuth, HomeViewData{
		User: &model.User{
			Username: "template_tester",
		},
		Categories:       []CategoryGroup{authCategory},
		TotalTasks:       1,
		CompletedTasks:   1,
		PercentCompleted: 100,
	})
	if err != nil {
		t.Fatalf("RenderHome (authenticated) failed: %v", err)
	}
	outputAuth := bufAuth.String()
	if !strings.Contains(outputAuth, "template_tester") {
		t.Errorf("expected home page to display username 'template_tester'")
	}
	if !strings.Contains(outputAuth, "Hello World") {
		t.Errorf("expected home page to display task title 'Hello World'")
	}
}

func TestTemplateRegistry_RenderEditor(t *testing.T) {
	registry, err := NewTemplateRegistry("")
	if err != nil {
		t.Fatalf("failed to create template registry: %v", err)
	}

	var buf bytes.Buffer
	err = registry.RenderEditor(&buf, EditorViewData{
		User: &model.User{Username: "editor_user"},
		Task: model.Task{
			ID:          42,
			Title:       "Test Problem",
			Category:    "Strings",
			Difficulty:  "easy",
			Objective:   "Reverse the string correctly.",
			Code:        "package main\n\nfunc Solution() {}",
		},
		PrevTaskID: 41,
		NextTaskID: 43,
	})
	if err != nil {
		t.Fatalf("RenderEditor failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Problem") {
		t.Errorf("expected editor output to contain task title 'Test Problem'")
	}
	if !strings.Contains(output, "Reverse the string correctly.") {
		t.Errorf("expected editor output to contain objective")
	}
	if !strings.Contains(output, "/problems/41") {
		t.Errorf("expected editor output to link to prev problem 41")
	}
	if !strings.Contains(output, "/problems/43") {
		t.Errorf("expected editor output to link to next problem 43")
	}
	if !strings.Contains(output, "editor_user") {
		t.Errorf("expected editor output to display username in header")
	}
}

func TestTemplateRegistry_RenderSignupPage(t *testing.T) {
	registry, err := NewTemplateRegistry("")
	if err != nil {
		t.Fatalf("failed to create template registry: %v", err)
	}

	var buf bytes.Buffer
	err = registry.RenderSignupPage(&buf, SignupPageViewData{
		User: &model.User{},
	})
	if err != nil {
		t.Fatalf("RenderSignupPage failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Create an account") {
		t.Errorf("expected signup page to contain 'Create an account'")
	}
	if !strings.Contains(output, "Go-Solve") {
		t.Errorf("expected signup page to contain header brand")
	}
}

func TestTemplateRegistry_RenderPartials(t *testing.T) {
	registry, err := NewTemplateRegistry("")
	if err != nil {
		t.Fatalf("failed to create template registry: %v", err)
	}

	// 1. Signup Form Partial - Username Taken
	var bufSignupTaken bytes.Buffer
	err = registry.RenderSignupForm(&bufSignupTaken, SignupFormData{
		UsernameTaken: true,
	})
	if err != nil {
		t.Fatalf("RenderSignupForm failed: %v", err)
	}
	if !strings.Contains(bufSignupTaken.String(), "Username is already taken.") {
		t.Errorf("expected signup form partial to show 'Username is already taken.'")
	}

	// 2. Signup Form Partial - Success
	var bufSignupSuccess bytes.Buffer
	err = registry.RenderSignupForm(&bufSignupSuccess, SignupFormData{
		SignUpSuccess: true,
	})
	if err != nil {
		t.Fatalf("RenderSignupForm failed: %v", err)
	}
	if !strings.Contains(bufSignupSuccess.String(), "Signed up successfully!") {
		t.Errorf("expected signup form partial to show 'Signed up successfully!'")
	}

	// 3. Login Form Partial - Password Incorrect
	var bufLoginPw bytes.Buffer
	err = registry.RenderLoginForm(&bufLoginPw, LoginFormData{
		PasswordIncorrect: true,
	})
	if err != nil {
		t.Fatalf("RenderLoginForm failed: %v", err)
	}
	if !strings.Contains(bufLoginPw.String(), "Incorrect password.") {
		t.Errorf("expected login form partial to show 'Incorrect password.'")
	}

	// 4. Login Form Partial - Username Not Found
	var bufLoginUser bytes.Buffer
	err = registry.RenderLoginForm(&bufLoginUser, LoginFormData{
		UsernameNotFound: true,
	})
	if err != nil {
		t.Fatalf("RenderLoginForm failed: %v", err)
	}
	if !strings.Contains(bufLoginUser.String(), "Username not found.") {
		t.Errorf("expected login form partial to show 'Username not found.'")
	}
}
