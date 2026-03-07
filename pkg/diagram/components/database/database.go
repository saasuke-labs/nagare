package database

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const AllowedActions = "read,write"

const (
	DefaultWidth  = 200.0
	DefaultHeight = 200.0
)

//go:embed database.html
var templateFile embed.FS

var tmpl = template.Must(template.New("database").Funcs(template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"div": func(a, b float64) float64 { return a / b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
}).ParseFS(templateFile, "database.html"))

type Props struct {
	Title           string `prop:"title"`
	Engine          string `prop:"engine"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:           "Database",
		Engine:          "PostgreSQL",
		BackgroundColor: "#0f766e",
		ForegroundColor: "#ecfdf5",
		AccentColor:     "#14b8a6",
	}
}

type Legacy struct {
	components.Shape
	Text    string
	Props   Props
	State   string
	Actions map[string][]map[string]any
}

func NewLegacy(id string) *Legacy {

	return &Legacy{Text: id, Props: DefaultProps(), Actions: make(map[string][]map[string]any)}
}

type templateData struct {
	X       float64
	Y       float64
	Width   float64
	Height  float64
	Props   Props
	Text    string
	Actions map[string][]map[string]any
}

func (l *Legacy) data() templateData {
	return templateData{X: l.X, Y: l.Y, Width: l.Width, Height: l.Height, Props: l.Props, Text: l.Text, Actions: l.Actions}
}

func (l *Legacy) Draw() string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "database", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering database template: %v -->", err)
	}
	return buf.String()
}

func (l *Legacy) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Legacy) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := NewLegacy(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	_ = comp.Props.Parse(propsRaw(nodeProps))
	comp.Actions = actionMapFromAny(nodeProps["actions"])
	return comp.Draw()
}

func BuildLegacy(id string, parent *components.Shape) *Legacy {
	comp := NewLegacy(id)
	comp.Shape = components.Shape{Width: DefaultWidth, Height: DefaultHeight}
	if parent != nil {
		comp.Shape.X = parent.X + comp.Shape.X
		comp.Shape.Y = parent.Y + comp.Shape.Y
	}
	return comp
}

func propsRaw(nodeProps map[string]any) string {
	if v, ok := nodeProps["_rawProps"].(string); ok {
		return v
	}
	return ""
}

func actionMapFromAny(v any) map[string][]map[string]any {
	out := make(map[string][]map[string]any)
	typed, ok := v.(map[string][]map[string]any)
	if !ok {
		return out
	}
	for k, items := range typed {
		copied := make([]map[string]any, 0, len(items))
		for _, item := range items {
			dup := make(map[string]any, len(item))
			for key, val := range item {
				dup[key] = val
			}
			copied = append(copied, dup)
		}
		out[k] = copied
	}
	return out
}
