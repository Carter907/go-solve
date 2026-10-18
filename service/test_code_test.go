package service

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestTestSolution_Passing(t *testing.T) {
	code := `package hello_world

func ReturnMessage() string {
	return "Hello World"
}
`
	taskPath := filepath.Join("..", "tasks", "easy", "hello_world")
	res, err := TestSolution(code, taskPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Passed {
		t.Fatalf("expected solution to pass, got Out=%q Err=%q", res.Out, res.Err)
	}
	if res.CompErr {
		t.Fatalf("expected CompErr=false, got true")
	}
}

func TestTestSolution_FailingTests(t *testing.T) {
	code := `package hello_world

func ReturnMessage() string {
	return "Wrong Message"
}
`
	taskPath := filepath.Join("..", "tasks", "easy", "hello_world")
	res, err := TestSolution(code, taskPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Passed {
		t.Fatalf("expected solution to fail tests, but passed")
	}
	if res.CompErr {
		t.Fatalf("expected CompErr=false for logic error, got true")
	}
}

func TestTestSolution_CompilationError(t *testing.T) {
	code := `package hello_world

func ReturnMessage() string {
	this is a syntax error
}
`
	taskPath := filepath.Join("..", "tasks", "easy", "hello_world")
	res, err := TestSolution(code, taskPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Passed {
		t.Fatalf("expected compile failure, but passed")
	}
	if !res.CompErr {
		t.Fatalf("expected CompErr=true for syntax error, got false")
	}
}

func TestTestSolution_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	code := `package hello_world

import "time"

func ReturnMessage() string {
	time.Sleep(15 * time.Second)
	return "Hello World"
}
`
	taskPath := filepath.Join("..", "tasks", "easy", "hello_world")
	res, err := TestSolution(code, taskPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Passed {
		t.Fatalf("expected timeout failure, but passed")
	}
	if !strings.Contains(res.Err, "timed out") {
		t.Fatalf("expected Err to mention timeout, got %q", res.Err)
	}
}
