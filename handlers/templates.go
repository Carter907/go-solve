package handlers

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/Carter907/go-solve/model"
)

var funcMap = template.FuncMap{
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
}

type HomeViewData struct {
	User             *model.User
	Categories       []CategoryGroup
	TotalTasks       int
	CompletedTasks   int
	PercentCompleted int
}

type EditorViewData struct {
	User *model.User
	model.Task
	PrevTaskID uint
	NextTaskID uint
}

type SignupPageViewData struct {
	User *model.User
	SignupFormData
}

type SignupFormData struct {
	UsernameTaken      bool
	PasswordsDontMatch bool
	SignUpSuccess      bool
}

type LoginFormData struct {
	UsernameNotFound  bool
	PasswordIncorrect bool
}

type TemplateRegistry struct {
	home       *template.Template
	editor     *template.Template
	signupPage *template.Template
	signupForm *template.Template
	loginForm  *template.Template
}

func resolveTemplateDir(dir string) string {
	if dir != "" {
		return dir
	}
	if _, err := os.Stat("templates"); err == nil {
		return "templates"
	}
	if _, err := os.Stat("../templates"); err == nil {
		return "../templates"
	}
	return "templates"
}

func parseTemplateFiles(name string, dir string, filenames ...string) (*template.Template, error) {
	paths := make([]string, len(filenames))
	for i, f := range filenames {
		paths[i] = filepath.Join(dir, f)
	}
	t := template.New(name).Funcs(funcMap)
	return t.ParseFiles(paths...)
}

func NewTemplateRegistry(dir string) (*TemplateRegistry, error) {
	dir = resolveTemplateDir(dir)

	baseFiles := []string{
		"base.html",
		"header.html",
		"login-button.html",
		"logout-button.html",
		"login-form.html",
	}

	homeFiles := append(append([]string{}, baseFiles...), "task-list.html", "task.html")
	homeTmpl, err := parseTemplateFiles("base", dir, homeFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse home template: %w", err)
	}

	editorFiles := append(append([]string{}, baseFiles...), "editor.html")
	editorTmpl, err := parseTemplateFiles("base", dir, editorFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse editor template: %w", err)
	}

	signupPageFiles := append(append([]string{}, baseFiles...), "signup-form.html", "signup-page.html")
	signupPageTmpl, err := parseTemplateFiles("base", dir, signupPageFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signup page template: %w", err)
	}

	signupFormTmpl, err := parseTemplateFiles("signup-form", dir, "signup-form.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse signup form partial: %w", err)
	}

	loginFormTmpl, err := parseTemplateFiles("login-form", dir, "login-form.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse login form partial: %w", err)
	}

	return &TemplateRegistry{
		home:       homeTmpl,
		editor:     editorTmpl,
		signupPage: signupPageTmpl,
		signupForm: signupFormTmpl,
		loginForm:  loginFormTmpl,
	}, nil
}

var (
	templatesOnce sync.Once
	Templates     *TemplateRegistry
)

func InitTemplates() error {
	var err error
	templatesOnce.Do(func() {
		Templates, err = NewTemplateRegistry("")
	})
	return err
}

func GetTemplates() *TemplateRegistry {
	if Templates == nil {
		_ = InitTemplates()
	}
	return Templates
}

func (r *TemplateRegistry) RenderHome(w io.Writer, data HomeViewData) error {
	return r.home.ExecuteTemplate(w, "base", data)
}

func (r *TemplateRegistry) RenderEditor(w io.Writer, data EditorViewData) error {
	return r.editor.ExecuteTemplate(w, "base", data)
}

func (r *TemplateRegistry) RenderSignupPage(w io.Writer, data SignupPageViewData) error {
	return r.signupPage.ExecuteTemplate(w, "base", data)
}

func (r *TemplateRegistry) RenderSignupForm(w io.Writer, data SignupFormData) error {
	return r.signupForm.ExecuteTemplate(w, "signup-form", data)
}

func (r *TemplateRegistry) RenderLoginForm(w io.Writer, data LoginFormData) error {
	return r.loginForm.ExecuteTemplate(w, "login-form", data)
}
