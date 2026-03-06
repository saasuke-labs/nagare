package artifact

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/props"
)

//go:embed artifact.html
var templateFile embed.FS

var tmpl = template.Must(template.New("artifact").Funcs(template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
}).ParseFS(templateFile, "artifact.html"))

type Props struct {
	Title           string `prop:"title"`
	Filename        string `prop:"filename"`
	Size            string `prop:"size"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:           "Artifact",
		Filename:        "build.tar.gz",
		Size:            "24 MB",
		BackgroundColor: "#1f2937",
		ForegroundColor: "#f9fafb",
		AccentColor:     "#9ca3af",
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
	if err := tmpl.ExecuteTemplate(&buf, "artifact", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering artifact template: %v -->", err)
	}
	return buf.String()
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}
