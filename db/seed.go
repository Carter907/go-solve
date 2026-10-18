package db

import (
	"fmt"
	"log"

	"github.com/Carter907/go-solve/model"
	"gorm.io/gorm"
)

// SeedTasks populates the database with initial tasks if the task table is empty.
func SeedTasks(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Task{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count existing tasks: %w", err)
	}

	if count > 0 {
		return nil
	}

	defaultTasks := []model.Task{
		{
			Title:       "Hello World",
			Description: "Learn how to return a string in Go",
			Difficulty:  "easy",
			Code: `package hello_world

func ReturnMessage() string {

}
`,
			Objective: "Return \"Hello World\" to from this function",
		},
		{
			Title:       "Bubbles",
			Description: "Sort an integer slice with the Bubble Sort algorithm",
			Difficulty:  "easy",
			Code: `package bubbles

/* Please don't use the slices package
*  for this problem, and the purpose of the
*  task is to specifically use the Bubble Sorting Algorithm
**/

func Bubbles(data []int) {

}
`,
			Objective: "Sort the int slice using the Bubble Sorting algorithm.",
		},
	}

	for _, task := range defaultTasks {
		if err := db.Create(&task).Error; err != nil {
			return fmt.Errorf("failed to seed task %q: %w", task.Title, err)
		}
	}

	log.Printf("Successfully seeded %d default tasks into database", len(defaultTasks))
	return nil
}
