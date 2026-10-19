package model

import (
	"encoding/json"
	"fmt"
)

type Task struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `gorm:"not null" json:"description"`
	Difficulty  string `gorm:"not null" json:"difficulty"`
	Category    string `gorm:"not null;default:'Basics';index" json:"category"`
	Code        string `gorm:"not null" json:"code"`
	Objective   string `gorm:"not null" json:"objective"`
}

func (Task) TableName() string {
	return "task"
}

func (t Task) String() string {
	jsonData, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return string(jsonData)
}

type TaskResult struct {
	Out     string `json:"output_stream"`
	Err     string `json:"error_stream"`
	Passed  bool   `json:"tests_passed"`
	CompErr bool   `json:"compile_err"`
}

func (t TaskResult) String() string {
	jsonData, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return string(jsonData)
}
