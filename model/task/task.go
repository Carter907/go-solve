package task

type Task struct {
	Title string `json:"title"`
	Task  string `json:"task"`
	Code  string `json:"code"`
	Path  string `json:"test-path"`
}
