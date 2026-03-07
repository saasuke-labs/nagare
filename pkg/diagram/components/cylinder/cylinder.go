package cylinder

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	DefaultWidth  = 200.0
	DefaultHeight = 200.0
)

//go:embed cylinder.html
var templateFile embed.FS

var tmpl = template.Must(template.New("cylinder").Funcs(template.FuncMap{
	"div": func(a, b float64) float64 { return a / b },
}).ParseFS(templateFile, "cylinder.html"))

type Props struct {
	Title           string `prop:"title"`
	Subtitle        string `prop:"subtitle"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:           "Cylinder",
		Subtitle:        "",
		BackgroundColor: "#0f766e",
		ForegroundColor: "#ecfdf5",
		AccentColor:     "#14b8a6",
	}
}

type Legacy struct {
	components.Shape
	Text  string
	Props Props
}

func NewLegacy(id string) *Legacy {
	return &Legacy{Text: id, Props: DefaultProps()}
}

type templateData struct {
	ID     string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  Props
}

func (l *Legacy) data() templateData {
	return templateData{ID: l.Text, X: l.X, Y: l.Y, Width: l.Width, Height: l.Height, Props: l.Props}
}

func (l *Legacy) Draw() string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "cylinder", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering cylinder template: %v -->", err)
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
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
