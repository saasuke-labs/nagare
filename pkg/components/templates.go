package components

import (
	"bytes"
	"embed"
	"html/template"
)

var funcMap = template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
}

//go:embed templates/*.html
var templateFiles embed.FS

var templates *template.Template

func init() {
	var err error

	// Create a new template and parse the embedded files
	templates = template.New("").Funcs(funcMap)
	templates, err = templates.ParseFS(templateFiles, "templates/*.html")

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
