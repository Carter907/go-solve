package handlers

import (
	"github.com/Carter907/go-solve/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBaseHandler(t *testing.T) {

	req, err := http.NewRequest("GET", "https://localhost:8080/", nil)
	if err != nil {
		t.Fatal(err)
	}

	tasks := make([]model.Task, 0)
	user := &model.User{
		ID:       0,
		Username: "",
		Password: "",
	}

	res := httptest.NewRecorder()
	BaseHandler(res, req, tasks, user)

}

func TestEditorHandler(t *testing.T) {

	_, err := http.NewRequest("GET", "https://localhost:8080/editor", nil)
	if err != nil {
		t.Fatal(err)
	}
}
