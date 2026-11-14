package terminal

import (
	"math"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	DefaultWidth  = 360.0
	DefaultHeight = 220.0
)

// Props are the configurable properties. Colour props are optional
// overrides; when empty the diagram theme decides the colours.
type Props struct {
	Title           string `prop:"title"`
	WorkingDir      string `prop:"cwd"`
	Command         string `prop:"command"`
	PromptSymbol    string `prop:"prompt"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{Title: "Terminal", WorkingDir: "~/project", Command: "go run main.go", PromptSymbol: "$"}
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
	t := c.Theme
	w, h := c.Box.Width, c.Box.Height
	barH := style.Clamp(h*0.16, 14, 28)
	inset := style.Clamp(barH*0.2, 3, 6)
	bodyY := barH
	bodyH := h - barH - inset
	body := style.Or(c.Props.BackgroundColor, t.CodeBg)
	fg := style.Legible(c.Props.ForegroundColor, body, t.CodeFg)
	accent := style.Or(c.Props.AccentColor, style.AccentTeal)

	var b strings.Builder
	b.WriteString(style.Window(style.WindowSpec{
		X: c.Box.X, Y: c.Box.Y, W: w, H: h,
		Class: "nagare-terminal", BarH: barH,
		CX: inset, CY: bodyY, CW: w - 2*inset, CH: bodyH,
		Title: c.Props.Title, TitleMono: true,
		ContentFill: body, Theme: t,
	}))

	size := style.Clamp(bodyH*0.13, 8, 13)
	pad := style.Clamp(size, 6, 14)
	textX := inset + pad
	textW := w - 2*(inset+pad)
	prompt := strings.TrimSpace(c.Props.PromptSymbol + " " + c.Props.Command)
	line := size * 1.6
	if c.Props.WorkingDir != "" && bodyH >= 2*line+pad {
		b.WriteString(style.Text(c.Props.WorkingDir, style.TextOpts{X: textX, Y: bodyY + pad + size/2, Size: size, Color: accent, Mono: true, Width: textW}))
		b.WriteString(style.Text(prompt, style.TextOpts{X: textX, Y: bodyY + pad + size/2 + line, Size: size, Color: fg, Mono: true, Width: textW}))
	} else if bodyH >= size+4 {
		b.WriteString(style.Text(prompt, style.TextOpts{X: textX, Y: bodyY + math.Min(pad+size/2, bodyH/2), Size: size, Color: fg, Mono: true, Width: textW}))
	}
	b.WriteString(style.WindowEnd)
	return b.String(), nil
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
