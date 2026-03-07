package apigateway

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/props"
)

//go:embed apigateway.html
var templateFile embed.FS

var tmpl = template.Must(template.New("api-gateway").Funcs(template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
}).ParseFS(templateFile, "apigateway.html"))

type Props struct {
	Title           string `prop:"title"`
	Route           string `prop:"route"`
	Method          string `prop:"method"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:           "API Gateway",
		Route:           "/api",
		Method:          "ANY",
		BackgroundColor: "#312e81",
		ForegroundColor: "#eef2ff",
		AccentColor:     "#6366f1",
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
	if err := tmpl.ExecuteTemplate(&buf, "api-gateway", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering API gateway template: %v -->", err)
	}
	return buf.String()
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}

const (
	DefaultWidth  = 220.0
	DefaultHeight = 180.0
)

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := NewLegacy(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
