package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Carter907/go-solve/model"
)

func RunCode(task *model.Task) model.TaskResult {
	path := fmt.Sprintf("./tasks/%s/%s",
		strings.ToLower(task.Difficulty),
		strings.ReplaceAll(
			strings.ToLower(task.Title), " ", "_"),
	)
	res, err := TestSolution(task.Code, path)
	if err != nil {
		log.Printf("Error running test code: %v\n", err)
		return model.TaskResult{
			Err:     fmt.Sprintf("Internal error: %v", err),
			Passed:  false,
			CompErr: true,
		}
	}

	return res
}

func TestSolution(code string, taskPath string) (model.TaskResult, error) {
	// 1. Create an isolated temporary directory
	tmpDir, err := os.MkdirTemp("", "gosolve-eval-*")
	if err != nil {
		return model.TaskResult{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 2. Write minimal go.mod for isolated execution
	goModContent := "module task\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o644); err != nil {
		return model.TaskResult{}, fmt.Errorf("failed to write go.mod: %w", err)
	}

	// 3. Copy only the task's test files (*_test.go) into the temp directory
	entries, err := os.ReadDir(taskPath)
	if err != nil {
		return model.TaskResult{}, fmt.Errorf("failed to read task directory %q: %w", taskPath, err)
	}

	testFileFound := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			testFileFound = true
			content, err := os.ReadFile(filepath.Join(taskPath, entry.Name()))
			if err != nil {
				return model.TaskResult{}, fmt.Errorf("failed to read test file %q: %w", entry.Name(), err)
			}
			if err := os.WriteFile(filepath.Join(tmpDir, entry.Name()), content, 0o644); err != nil {
				return model.TaskResult{}, fmt.Errorf("failed to copy test file %q: %w", entry.Name(), err)
			}
		}
	}

	if !testFileFound {
		return model.TaskResult{}, fmt.Errorf("no test files found in %q", taskPath)
	}

	// 4. Write user submission into the temp directory
	solutionPath := filepath.Join(tmpDir, "solution.go")
	if err := os.WriteFile(solutionPath, []byte(code), 0o644); err != nil {
		return model.TaskResult{}, fmt.Errorf("failed to write solution file: %w", err)
	}

	// 5. Execute go test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "-vet=off", "-v", ".")
	cmd.Dir = tmpDir

	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"GOTOOLCHAIN=local",
		"GOPROXY=off",
		"CGO_ENABLED=0",
	}
	if home := os.Getenv("HOME"); home != "" {
		env = append(env, "HOME="+home)
	}
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		env = append(env, "GOPATH="+gopath)
	}
	if gocache := os.Getenv("GOCACHE"); gocache != "" {
		env = append(env, "GOCACHE="+gocache)
	}
	if tmpdir := os.Getenv("TMPDIR"); tmpdir != "" {
		env = append(env, "TMPDIR="+tmpdir)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	if ctx.Err() == context.DeadlineExceeded {
		return model.TaskResult{
			Err:     "Execution timed out (10s limit exceeded). Check for infinite loops or blocking calls.",
			Passed:  false,
			CompErr: false,
		}, nil
	}

	if err != nil {
		compErr := strings.Contains(outputStr, "build failed") ||
			strings.Contains(outputStr, "syntax error") ||
			strings.Contains(outputStr, "undefined:") ||
			strings.Contains(outputStr, "cannot use")
		return model.TaskResult{
			Out:     outputStr,
			Passed:  false,
			CompErr: compErr,
		}, nil
	}

	return model.TaskResult{
		Out:     outputStr,
		Passed:  true,
		CompErr: false,
	}, nil
}
