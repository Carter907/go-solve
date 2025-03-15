package model

import (
	"encoding/json"
	"fmt"
)

type Task struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	Code        string `json:"code"`
	Objective   string `json:"objective"`
}

func (t Task) String() string {

	jsonData, err := json.MarshalIndent(t, "", "	")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return fmt.Sprintf("%s", string(jsonData))
}

type TaskResult struct {
	Out     string `json:"output_stream"`
	Err     string `json:"error_stream"`
	Passed  bool   `json:"tests_passed"`
	CompErr bool   `json:"compile_err"`
}

func (t TaskResult) String() string {
	jsonData, err := json.MarshalIndent(t, "", "	")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return fmt.Sprintf("%s", string(jsonData))
}
