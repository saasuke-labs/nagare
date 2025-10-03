package components

import (
	"bytes"
	"html/template"
	"path/filepath"
	"runtime"
)

var funcMap = template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
}

var templates *template.Template

func init() {
	var err error
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "templates", "*.html")

	// Create a new template and parse the files
	templates = template.New("").Funcs(funcMap)
	templates, err = templates.ParseGlob(dir)

	if err != nil {
		panic(err)
	}
}

func RenderTemplate(name string, data interface{}) (string, error) {
	var buf bytes.Buffer

	err := templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
