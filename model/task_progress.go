package model

import (
	"encoding/json"
	"fmt"
)

type TaskProgress struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	TaskID   uint   `json:"task_id"`
	Progress string `json:"progress"`
}

func (tp TaskProgress) String() string {

	jsonData, err := json.MarshalIndent(tp, "", "	")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return fmt.Sprintf("%s", string(jsonData))
}
