package model

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
}

func (User) TableName() string {
	return "user"
}

func (u User) String() string {
	jsonData, err := json.MarshalIndent(u, "", "\t")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return string(jsonData)
}
