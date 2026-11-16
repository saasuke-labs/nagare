package cylinder

import (
	"fmt"
	"math"
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
	if l.Theme.IsWatercolor() {
		return l.drawPainted()
	}
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

// drawPainted paints the cylinder: a wash for the body, a lighter one for
// the lid, a little shadow down one side so it turns, and the pen drawing the
// lid as an ellipse and the body as one stroke down, round and up.
func (l *Component) drawPainted() string {
	t := l.Theme
	w, h := l.Width, l.Height
	half := style.StrokeWidth / 2
	ry := style.Clamp(h*CapRatio, 4, 22)
	left, right := half, w-half
	top, bottom := ry+half, h-ry-half
	rx := (right - left) / 2
	cx := w / 2

	pigment, density := style.SurfacePaint(l.Props.BackgroundColor, style.Or(l.Props.AccentColor, style.AccentTeal), t)
	lidPigment := style.Pigment(style.Or(l.Props.AccentColor, style.AccentTeal))
	fill := style.PaintedColor(pigment, density, t)
	titleColor, mutedColor, mutedOpacity := style.TextColors(l.Props.ForegroundColor, fill, t)

	arc := func(from, to float64, n int) []style.Pt {
		pts := make([]style.Pt, 0, n+1)
		for i := 0; i <= n; i++ {
			a := from + (to-from)*float64(i)/float64(n)
			pts = append(pts, style.Pt{X: cx + math.Cos(a)*rx, Y: math.Sin(a) * ry})
		}
		return pts
	}
	// Body: down the left side, round the bottom, up the right, and back
	// along the front of the lid.
	var body []style.Pt
	for _, p := range arc(math.Pi, 0, 16) {
		body = append(body, style.Pt{X: p.X, Y: p.Y + bottom})
	}
	body = append([]style.Pt{{X: left, Y: top}}, body...)
	body = append(body, style.Pt{X: right, Y: top})
	for _, p := range arc(0, math.Pi, 16) {
		body = append(body, style.Pt{X: p.X, Y: p.Y + top})
	}
	// The shadow side: the right third of the body.
	var shade []style.Pt
	for _, p := range arc(math.Pi*0.32, 0, 6) {
		shade = append(shade, style.Pt{X: p.X, Y: p.Y + bottom})
	}
	shade = append(shade, style.Pt{X: right, Y: top})
	for _, p := range arc(0, math.Pi*0.32, 6) {
		shade = append(shade, style.Pt{X: p.X, Y: p.Y + top})
	}

	seed := t.Seed
	maxOff := style.Clamp(w*0.03, 1.5, 5)
	line := style.LineColor(pigment, t)
	ink := style.InkSpec{Width: 1.2, Color: line, Seed: seed + 1}

	var b strings.Builder
	fmt.Fprintf(&b, `<g class="nagare-cylinder" transform="translate(%s, %s)">`, style.F(l.X), style.F(l.Y))
	b.WriteString(style.Wash(style.WashSpec{Poly: body, Pigment: pigment, Density: density * 0.8, Seed: seed, Layers: 4, MaxOff: maxOff, Settle: 0.5}, t))
	b.WriteString(style.Wash(style.WashSpec{Poly: shade, Pigment: pigment, Density: density * 0.38, Seed: seed + 2, Layers: 3, MaxOff: maxOff}, t))
	b.WriteString(style.Wash(style.WashSpec{Poly: style.Ellipse(cx, top, rx, ry, 28), Pigment: lidPigment, Density: density * 0.45, Seed: seed + 3, Layers: 4, MaxOff: maxOff * 0.6}, t))

	b.WriteString(`<g class="nagare-ink" filter="url(#nagare-wc-ink)">`)
	b.WriteString(style.InkLoop(style.Ellipse(cx, top, rx, ry, 40), style.InkSpec{Width: 1.2, Color: line, Seed: seed + 4, NoLift: true}))
	var outline []style.Pt
	outline = append(outline, style.Pt{X: left, Y: top})
	for _, p := range arc(math.Pi, 0, 24) {
		outline = append(outline, style.Pt{X: p.X, Y: p.Y + bottom})
	}
	outline = append(outline, style.Pt{X: right, Y: top})
	ink.TaperIn, ink.TaperOut = 6, 6
	b.WriteString(style.InkStroke(outline, ink))
	b.WriteString(`</g>`)

	base := math.Min(w, h)
	titleSize := style.Clamp(base*0.1, 10, 17)
	subSize := style.Clamp(titleSize*0.78, 8.5, 12.5)
	textW := w - 2*style.Clamp(w*0.1, 4, 14)
	ty := (top + ry + h - ry*0.5) / 2
	if l.Props.Subtitle != "" && h >= 70 {
		gap := subSize * 0.45
		y := ty - (titleSize+gap+subSize)/2
		b.WriteString(style.Text(l.Props.Title, style.TextOpts{X: w / 2, Y: y + titleSize/2, Size: titleSize, Color: titleColor, Weight: "600", Anchor: "middle", Width: textW}))
		b.WriteString(style.Text(l.Props.Subtitle, style.TextOpts{X: w / 2, Y: y + titleSize + gap + subSize/2, Size: subSize, Color: mutedColor, Opacity: mutedOpacity, Anchor: "middle", Width: textW}))
	} else {
		b.WriteString(style.Text(l.Props.Title, style.TextOpts{X: w / 2, Y: ty, Size: titleSize, Color: titleColor, Weight: "600", Anchor: "middle", Width: textW}))
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
