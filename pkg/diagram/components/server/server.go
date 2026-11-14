package server

import (
	"fmt"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	DefaultWidth  = 200.0
	DefaultHeight = 140.0
)

// Props are the configurable properties. Colour props are optional
// overrides; when empty the diagram theme decides the colours.
type Props struct {
	Title           string `prop:"title"`
	Icon            string `prop:"icon"`
	Port            int    `prop:"port"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{Icon: "default", Port: 80}
}

type Component struct {
	IDVal     string
	Box       core.Shape
	Container core.Shape
	Props     Props
	Theme     style.Theme
}

func New(id string) *Component {
	return &Component{IDVal: id, Props: DefaultProps(), Theme: style.Light}
}

func (c *Component) ApplyProps(raw string) error { return c.Props.Parse(raw) }

func (c *Component) SetContainer(box core.Shape) { c.Container = box }

func (c *Component) SetShape(box core.Shape) { c.Box = box }

func (c *Component) Shape() core.Shape { return c.Box }

func (c *Component) Port(name string) (core.Point, error) {
	return core.RectPort(c.Box, name)
}

func (c *Component) Draw() (string, error) {
	title := style.Or(c.Props.Title, c.IDVal)
	sub := fmt.Sprintf(":%d", c.Props.Port)
	if kind := strings.TrimSpace(c.Props.Icon); kind != "" && kind != "default" {
		sub = kind + " · " + sub
	}
	return style.Card(style.CardSpec{
		X: c.Box.X, Y: c.Box.Y, W: c.Box.Width, H: c.Box.Height,
		Class:    "nagare-server",
		Title:    title,
		Subtitle: sub,
		Icon:     style.IconServer,
		Accent:   style.Or(c.Props.AccentColor, style.AccentBlue),
		Bg:       c.Props.BackgroundColor,
		Fg:       c.Props.ForegroundColor,
		Theme:    c.Theme,
	}), nil
}

type Adapter struct {
	Text  string
	Comp  *Component
	State string
}

func NewAdapter(id string) *Adapter {
	return &Adapter{Text: id, Comp: New(id)}
}

func (l *Adapter) Draw() string {
	result, _ := l.Comp.Draw()
	return result
}

func (l *Adapter) SetShape(shape components.Shape) { l.Comp.SetShape(shape) }

func (l *Adapter) Offset(dx, dy float64) {
	box := l.Comp.Shape()
	box.X -= dx
	box.Y -= dy
	l.Comp.SetShape(box)
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := NewAdapter(id)
	comp.SetShape(core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight))
	comp.Comp.Theme = style.ThemeFromProps(nodeProps)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Comp.ApplyProps(raw)
	}
	return comp.Draw()
}
