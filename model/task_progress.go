package model

import (
	"encoding/json"
	"fmt"
)

type TaskProgress struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID   uint   `gorm:"not null;uniqueIndex:idx_user_task" json:"user_id"`
	TaskID   uint   `gorm:"not null;uniqueIndex:idx_user_task" json:"task_id"`
	Progress string `gorm:"not null" json:"progress"`
}

func (TaskProgress) TableName() string {
	return "task_progress"
}

func (tp TaskProgress) String() string {
	jsonData, err := json.MarshalIndent(tp, "", "\t")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return string(jsonData)
}
