package cylinder

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
	DefaultHeight = 200.0
)

// Props are the configurable properties. bg, fg and accent are optional
// overrides; when empty the diagram theme decides the colours.
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
	return Props{Title: "Cylinder"}
}

type Component struct {
	components.Shape
	Text  string
	Props Props
	Theme style.Theme
}

func New(id string) *Component {
	return &Component{Text: id, Props: DefaultProps(), Theme: style.Light}
}

// CapRatio is the height of the cylinder's elliptical cap relative to the
// component height. The body label is centred below it.
const CapRatio = 0.11

func (l *Component) Draw() string {
	t := l.Theme
	w, h := l.Width, l.Height
	half := style.StrokeWidth / 2
	ry := style.Clamp(h*CapRatio, 4, 22)
	left, right := half, w-half
	top, bottom := ry+half, h-ry-half
	rx := (right - left) / 2

	fill := style.Or(l.Props.BackgroundColor, t.Surface)
	accent := style.Or(l.Props.AccentColor, style.AccentTeal)
	titleColor, mutedColor, mutedOpacity := style.TextColors(l.Props.ForegroundColor, fill, t)

	var b strings.Builder
	fmt.Fprintf(&b, `<g class="nagare-cylinder" transform="translate(%s, %s)">`, style.F(l.X), style.F(l.Y))
	// Body: straight sides with a rounded bottom.
	fmt.Fprintf(&b, `<path d="M%s %s V%s A%s %s 0 0 0 %s %s V%s" fill="%s" stroke="%s" stroke-width="%s"/>`,
		style.F(left), style.F(top), style.F(bottom), style.F(rx), style.F(ry), style.F(right), style.F(bottom), style.F(top),
		fill, t.Border, style.F(style.StrokeWidth))
	// Cap, tinted with the accent.
	fmt.Fprintf(&b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="%s"/>`, style.F(w/2), style.F(top), style.F(rx), style.F(ry), fill)
	fmt.Fprintf(&b, `<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="%s" fill-opacity="0.16" stroke="%s" stroke-width="%s"/>`,
		style.F(w/2), style.F(top), style.F(rx), style.F(ry), accent, accent, style.F(style.StrokeWidth))

	base := w
	if h < base {
		base = h
	}
	titleSize := style.Clamp(base*0.1, 10, 17)
	subSize := style.Clamp(titleSize*0.78, 8.5, 12.5)
	textW := w - 2*style.Clamp(w*0.1, 4, 14)
	// Centre of the visible body, below the cap.
	cy := (top + ry + h - ry*0.5) / 2
	if l.Props.Subtitle != "" && h >= 70 {
		gap := subSize * 0.45
		y := cy - (titleSize+gap+subSize)/2
		b.WriteString(style.Text(l.Props.Title, style.TextOpts{X: w / 2, Y: y + titleSize/2, Size: titleSize, Color: titleColor, Weight: "600", Anchor: "middle", Width: textW}))
		b.WriteString(style.Text(l.Props.Subtitle, style.TextOpts{X: w / 2, Y: y + titleSize + gap + subSize/2, Size: subSize, Color: mutedColor, Opacity: mutedOpacity, Anchor: "middle", Width: textW}))
	} else {
		b.WriteString(style.Text(l.Props.Title, style.TextOpts{X: w / 2, Y: cy, Size: titleSize, Color: titleColor, Weight: "600", Anchor: "middle", Width: textW}))
	}
	b.WriteString(`</g>`)
	return b.String()
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
