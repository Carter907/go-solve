package task

import (
	"encoding/json"
	"fmt"
)

type Task struct {
	Title string `json:"title"`
	Task  string `json:"task"`
	Code  string `json:"code"`
	Path  string `json:"test-path"`
}

func (t Task) String() string {

	jsonData, err := json.MarshalIndent(t, "", "	")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return fmt.Sprintf("%s", string(jsonData))
}
