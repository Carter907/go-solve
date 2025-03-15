package model

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u User) String() string {

	jsonData, err := json.MarshalIndent(u, "", "	")
	if err != nil {
		fmt.Println("There was an error marshalling the data: ", err)
	}

	return fmt.Sprintf("%s", string(jsonData))
}
