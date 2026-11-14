package cdn

import (
	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	DefaultWidth  = 200.0
	DefaultHeight = 160.0
)

// Props are the configurable properties. bg, fg and accent are optional
// overrides; when empty the diagram theme decides the colours.
type Props struct {
	Title           string `prop:"title"`
	Provider        string `prop:"provider"`
	Region          string `prop:"region"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:    "Edge",
		Provider: "Cloudflare",
		Region:   "Global",
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
		Class:    "nagare-cdn",
		Title:    l.Props.Title,
		Subtitle: l.Props.Provider,
		Meta:     l.Props.Region,
		Icon:     style.IconGlobe,
		Accent:   style.Or(l.Props.AccentColor, style.AccentCyan),
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
