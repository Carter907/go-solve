package editor

import (
	"bytes"
	"github.com/Carter907/go-solve/model/task"
	"log"
	"os"
	"os/exec"
)

func RunCode(task *task.Task) (out, errOut bytes.Buffer) {
	return TestSolution(task.Code, task.Path)
}
func TestSolution(code string, path string) (out, errOut bytes.Buffer) {
	file, err := os.Create(path + "/solution.go")
	if err != nil {
		log.Fatalln("invalid path: ", err)
	}
	file.WriteString(code)
	file.Close()
	cmd := exec.Command("go", "build", path+"/solution.go")
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		return
	}

	cmd = exec.Command("go", "test", path)
	out.Reset()
	errOut.Reset()
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}
