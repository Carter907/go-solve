package main

import (
	"bytes"
	"fmt"
	"github.com/Carter907/go-solve/model"
	"log"
	"os"
	"os/exec"
)

func RunCode(task *model.Task) model.TaskResult {
	out, errOut := TestSolution(task.Code, task.Path)

	return model.TaskResult{
		Out: out.String(),
		Err: errOut.String(),
	}
}

func TestSolution(code string, path string) (out, errOut bytes.Buffer) {

	fmt.Println("testing solution...")
	fmt.Println("code --\n", code)
	fmt.Println("path --", path)

	file, err := os.Create(path + "/solution.go")

	if err != nil {
		log.Fatalln("invalid path: ", err)
	}

	_, err = file.WriteString(code)

	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}
	}

	err = file.Close()

	if err != nil {
		return
	}

	cmd := exec.Command("go", "build", path+"/solution.go") // test that the code compiles

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err = cmd.Run()

	if err != nil {
		fmt.Println("Failed to run go build:", errOut.String())
		return
	}

	cmd = exec.Command("go", "test", path) // test that the code passes the tests

	out.Reset()
	errOut.Reset()

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err = cmd.Run()

	if err != nil {
		fmt.Println("Failed to run go test:", errOut.String())
		return
	}

	fmt.Println("Tests Passed!")

	return
}
