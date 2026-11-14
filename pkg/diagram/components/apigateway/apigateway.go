package apigateway

import (
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	DefaultWidth  = 220.0
	DefaultHeight = 180.0
)

// Props are the configurable properties. bg, fg and accent are optional
// overrides; when empty the diagram theme decides the colours.
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
		Title:  "API Gateway",
		Route:  "/api",
		Method: "ANY",
	}
}

type Component struct {
	components.Shape
	Text  string
	Props Props
	State string
	Theme style.Theme
}

func New(id string) *Component {
	return &Component{Text: id, Props: DefaultProps(), Theme: style.Light}
}

func (l *Component) Draw() string {
	return style.Card(style.CardSpec{
		X: l.X, Y: l.Y, W: l.Width, H: l.Height,
		Class:    "nagare-apigateway",
		Title:    l.Props.Title,
		Subtitle: strings.TrimSpace(l.Props.Method + " " + l.Props.Route),
		Meta:     "",
		Icon:     style.IconGateway,
		Accent:   style.Or(l.Props.AccentColor, style.AccentViolet),
		Bg:       l.Props.BackgroundColor,
		Fg:       l.Props.ForegroundColor,
		Theme:    l.Theme,
	})
}

func (l *Component) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Component) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := New(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	comp.Theme = style.ThemeFromProps(nodeProps)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	return comp.Draw()
}
