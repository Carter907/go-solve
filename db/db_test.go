package db

import (
	"testing"
)

func TestGetAllTasks(t *testing.T) {
	_, err := GetAllTasks()
	if err != nil {
		t.Fatal(err)
	}
}
