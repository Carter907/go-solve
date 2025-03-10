package editor

import (
	"bytes"
	"fmt"
	"github.com/Carter907/go-solve/model/task"
	"log"
	"os"
	"os/exec"
)

func RunCode(task *task.Task) (out, errOut bytes.Buffer) {
	return TestSolution(task.Code, task.Path)
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
		return bytes.Buffer{}, bytes.Buffer{}
	}
	cmd := exec.Command("go", "build", path+"/solution.go")
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to run go build:", errOut.String())
		return
	}

	cmd = exec.Command("go", "test", path)
	out.Reset()
	errOut.Reset()
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to run go test:", errOut.String())
		return
	}

	fmt.Println("console results --")
	fmt.Printf("out: %s\n", out.String())
	fmt.Printf("err: %v\n", errOut.String())

	return
}
