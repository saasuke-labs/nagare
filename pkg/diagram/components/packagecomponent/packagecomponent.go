package packagecomponent

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/props"
)

//go:embed packagecomponent.html
var templateFile embed.FS

var tmpl = template.Must(template.New("package").Funcs(template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
}).ParseFS(templateFile, "packagecomponent.html"))

type Props struct {
	Title           string `prop:"title"`
	Version         string `prop:"version"`
	Language        string `prop:"lang"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:           "Package",
		Version:         "1.0.0",
		Language:        "Go",
		BackgroundColor: "#92400e",
		ForegroundColor: "#fef3c7",
		AccentColor:     "#f97316",
	}
}

type Legacy struct {
	components.Shape
	Text  string
	Props Props
	State string
}

func NewLegacy(id string) *Legacy {
	return &Legacy{Text: id, Props: DefaultProps()}
}

type templateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  Props
	Text   string
}

func (l *Legacy) data() templateData {
	return templateData{X: l.X, Y: l.Y, Width: l.Width, Height: l.Height, Props: l.Props, Text: l.Text}
}

func (l *Legacy) Draw() string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "package", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering package template: %v -->", err)
	}
	return buf.String()
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}
