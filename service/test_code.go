package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Carter907/go-solve/model"
)

func RunCode(task *model.Task) model.TaskResult {
	path := fmt.Sprintf("./tasks/%s/%s",
		strings.ToLower(task.Difficulty),
		strings.ReplaceAll(
			strings.ToLower(task.Title), " ", "_"),
	)
	fmt.Println(path)
	res, err := TestSolution(task.Code, path)
	if err != nil {
		log.Fatalf("Failed to run test code: %v\n", err)
	}

	return res
}

func TestSolution(code string, path string) (model.TaskResult, error) {

	fmt.Println("testing solution...")
	fmt.Println("code --\n", code)
	fmt.Println("path --", path)

	file, err := os.Create(path + "/solution.go")
	if err != nil {
		log.Fatalf("Failed to create solution file: %v\n", err)
	}

	_, err = file.WriteString(code)
	if err != nil {
		log.Fatalf("Error writing to solution file: %v\n", err)
	}

	err = file.Close()
	if err != nil {
		log.Fatalf("Error closing solution file: %v\n", err)
	}

	cmd := exec.Command("go", "build", path+"/solution.go") // test that the code compiles

	out := bytes.Buffer{}
	errOut := bytes.Buffer{}

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err = cmd.Run()
	if err != nil {
		log.Printf("Error running build command: %v\n", err)
		return model.TaskResult{
			Out:     out.String(),
			Err:     errOut.String(),
			Passed:  false,
			CompErr: true,
		}, nil
	}

	cmd = exec.Command("go", "test", path) // test that the code passes the tests

	out.Reset()
	errOut.Reset()

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err = cmd.Run()
	if err != nil {
		log.Printf("Error running tests: %v\n", err)
		return model.TaskResult{
			Out:     out.String(),
			Err:     errOut.String(),
			Passed:  false,
			CompErr: false,
		}, nil
	}

	fmt.Println("Tests Passed!")

	return model.TaskResult{
		Out:     out.String(),
		Err:     errOut.String(),
		Passed:  true,
		CompErr: false,
	}, nil
}
