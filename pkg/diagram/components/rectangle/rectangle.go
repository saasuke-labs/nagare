package rectangle

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/props"
)

//go:embed rectangle.html
var templateFile embed.FS

var tmpl = template.Must(template.New("rectangle").Funcs(template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
}).ParseFS(templateFile, "rectangle.html"))

type Props struct {
	Title           string `prop:"title"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		BackgroundColor: "#e6f3ff",
		ForegroundColor: "#333333",
	}
}

type Component struct {
	IDVal     string
	Box       core.BoundingBox
	Container core.BoundingBox
	Props     Props
}

func New(id string) *Component {
	return &Component{IDVal: id, Props: DefaultProps()}
}

func (c *Component) ApplyProps(raw string) error {
	return c.Props.Parse(raw)
}

func (c *Component) SetContainer(box core.BoundingBox) {
	c.Container = box
}

func (c *Component) SetBoundingBox(box core.BoundingBox) {
	c.Box = box
}

func (c *Component) BoundingBox() core.BoundingBox {
	return c.Box
}

func (c *Component) Port(name string) (core.Point, error) {
	return core.RectPort(c.Box, name)
}

func (c *Component) Draw() (string, error) {
	displayText := c.IDVal
	if c.Props.Title != "" {
		displayText = c.Props.Title
	}

	data := struct {
		X             float64
		Y             float64
		Width         float64
		Height        float64
		DisplayText   string
		Background    string
		Foreground    string
		BorderRadiusX float64
		BorderRadiusY float64
	}{
		X:             c.Box.X,
		Y:             c.Box.Y,
		Width:         c.Box.Width,
		Height:        c.Box.Height,
		DisplayText:   displayText,
		Background:    c.Props.BackgroundColor,
		Foreground:    c.Props.ForegroundColor,
		BorderRadiusX: c.Box.Height * 0.1,
		BorderRadiusY: c.Box.Height * 0.1,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "rectangle", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type Legacy struct {
	Text  string
	Comp  *Component
	State string
}

func NewLegacy(id string) *Legacy {
	return &Legacy{Text: id, Comp: New(id)}
}

func (l *Legacy) Draw() string {
	result, err := l.Comp.Draw()
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering rectangle template: %v -->", err)
	}
	return result
}

func (l *Legacy) SetShape(shape components.Shape) {
	l.Comp.SetBoundingBox(core.BoundingBox{X: shape.X, Y: shape.Y, Width: shape.Width, Height: shape.Height})
}

func (l *Legacy) Offset(dx, dy float64) {
	box := l.Comp.BoundingBox()
	box.X -= dx
	box.Y -= dy
	l.Comp.SetBoundingBox(box)
}
