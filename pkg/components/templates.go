package components

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

var funcMap = template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
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
	fmt.Println("Rendering template: ", name, data)
	err := templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		fmt.Println("Error executing template: ", err)
		return "", err
	}
	fmt.Println("Returning result", buf.String())
	return buf.String(), nil
}
