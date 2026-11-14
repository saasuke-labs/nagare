package backgroundworker

import (
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
	Job             string `prop:"job"`
	Schedule        string `prop:"schedule"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{
		Title:    "Worker",
		Job:      "process-emails",
		Schedule: "@every 1m",
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
		Class:    "nagare-backgroundworker",
		Title:    l.Props.Title,
		Subtitle: l.Props.Job,
		Meta:     l.Props.Schedule,
		Icon:     style.IconClock,
		Accent:   style.Or(l.Props.AccentColor, style.AccentPink),
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
