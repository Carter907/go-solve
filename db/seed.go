package db

import (
	"errors"
	"fmt"
	"log"

	"github.com/Carter907/go-solve/model"
	"gorm.io/gorm"
)

// SeedTasks populates the database with default tasks if they are not already present,
// and ensures existing tasks have their Category properly assigned.
func SeedTasks(db *gorm.DB) error {
	defaultTasks := []model.Task{
		{
			Title:       "Hello World",
			Description: "Learn how to return a string in Go",
			Difficulty:  "easy",
			Category:    "Basics",
			Code: `package hello_world

func ReturnMessage() string {

}
`,
			Objective: "Return \"Hello World\" from this function.",
		},
		{
			Title:       "Sum Two",
			Description: "Return the sum of two integers",
			Difficulty:  "easy",
			Category:    "Basics",
			Code: `package sum_two

func SumTwo(a, b int) int {

}
`,
			Objective: "Write a function SumTwo(a, b int) int that returns the sum of a and b.",
		},
		{
			Title:       "Reverse String",
			Description: "Reverse a given string",
			Difficulty:  "easy",
			Category:    "Strings",
			Code: `package reverse_string

func ReverseString(s string) string {

}
`,
			Objective: "Write a function ReverseString(s string) string that returns the reversed string.",
		},
		{
			Title:       "Palindrome Check",
			Description: "Check if a string is a palindrome",
			Difficulty:  "easy",
			Category:    "Strings",
			Code: `package palindrome_check

func IsPalindrome(s string) bool {

}
`,
			Objective: "Write a function IsPalindrome(s string) bool that returns true if s reads identically forwards and backwards.",
		},
		{
			Title:       "Bubbles",
			Description: "Sort an integer slice with the Bubble Sort algorithm",
			Difficulty:  "easy",
			Category:    "Slices & Sorting",
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
		{
			Title:       "Find Maximum",
			Description: "Find the maximum integer in a slice",
			Difficulty:  "easy",
			Category:    "Slices & Sorting",
			Code: `package find_maximum

func FindMax(nums []int) int {

}
`,
			Objective: "Write a function FindMax(nums []int) int that returns the largest value in the slice.",
		},
		{
			Title:       "Word Count",
			Description: "Count occurrences of each word in a string",
			Difficulty:  "medium",
			Category:    "Maps & Structs",
			Code: `package word_count

func WordCount(s string) map[string]int {

}
`,
			Objective: "Write a function WordCount(s string) map[string]int returning a map of words to their frequencies.",
		},
	}

	seededCount := 0
	for _, task := range defaultTasks {
		var existing model.Task
		err := db.Where("title = ?", task.Title).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&task).Error; err != nil {
					return fmt.Errorf("failed to seed task %q: %w", task.Title, err)
				}
				seededCount++
			} else {
				return fmt.Errorf("failed to query task %q: %w", task.Title, err)
			}
		} else {
			if existing.Category != task.Category {
				if err := db.Model(&existing).Update("category", task.Category).Error; err != nil {
					return fmt.Errorf("failed to update category for task %q: %w", task.Title, err)
				}
			}
		}
	}

	if seededCount > 0 {
		log.Printf("Successfully seeded %d new default tasks into database", seededCount)
	}
	return nil
}
