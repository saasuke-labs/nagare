package led

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const AllowedActions = "blink"

const (
	DefaultWidth  = 20.0
	DefaultHeight = 20.0
)

//go:embed led.html
var templateFile embed.FS

var tmpl = template.Must(template.New("led").Funcs(template.FuncMap{
	"div": func(a, b float64) float64 { return a / b },
}).ParseFS(templateFile, "led.html"))

type Props struct {
	Mode string `prop:"mode"`
}

func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

func DefaultProps() Props {
	return Props{Mode: "green"}
}

type Component struct {
	components.Shape
	Text    string
	Props   Props
	Actions map[string][]map[string]any
	Theme   style.Theme
}

func New(id string) *Component {
	return &Component{Text: id, Props: DefaultProps(), Actions: map[string][]map[string]any{}}
}

type templateData struct {
	X               float64
	Y               float64
	Width           float64
	Height          float64
	BaseColor       string
	PulseColors     string
	InactiveOpacity string
	Actions         map[string][]map[string]any
}

func (l *Component) data() templateData {
	baseColor, pulseColors, inactiveOpacity := palette(l.Props.Mode)
	return templateData{
		X:               l.X,
		Y:               l.Y,
		Width:           l.Width,
		Height:          l.Height,
		BaseColor:       baseColor,
		PulseColors:     pulseColors,
		InactiveOpacity: inactiveOpacity,
		Actions:         l.Actions,
	}
}

func (l *Component) Draw() string {
	if l.Theme.IsWatercolor() {
		return l.drawPainted()
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "led", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering led template: %v -->", err)
	}
	return buf.String()
}

func (l *Component) SetShape(shape components.Shape) { l.Shape = shape }

func (l *Component) Offset(dx, dy float64) {
	l.X -= dx
	l.Y -= dy
}

func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := New(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	comp.Actions = actionMapFromAny(nodeProps["actions"])
	comp.Theme = style.ThemeFromProps(nodeProps)
	return comp.Draw()
}

func palette(mode string) (baseColor string, pulseColors string, inactiveOpacity string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "red":
		return "darkred", "darkred;#f87171;#ef4444;#dc2626;#b91c1c;#f87171;darkred", "0"
	default:
		return "darkgreen", "darkgreen;#4ade80;#22c55e;#86efac;#16a34a;#4ade80;darkgreen", "1"
	}
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

// drawPainted paints the light as a dot of pigment. It blinks the same way
// the flat light does: the dot's colour pulses, and a red light that is off
// at rest only shows while it blinks.
func (l *Component) drawPainted() string {
	d := l.data()
	base, pulses := "#3a7a30", d.PulseColors
	if strings.EqualFold(strings.TrimSpace(l.Props.Mode), "red") {
		base = "#9e2b2f"
	}
	pulses = strings.Replace(strings.Replace(pulses, "darkgreen", base, -1), "darkred", base, -1)
	r := l.Width / 2
	poly := style.Ellipse(r, l.Height/2, r, l.Height/2, 12)
	wash := style.Wash(style.WashSpec{Poly: poly, Pigment: base, Density: 0.78, Seed: l.Theme.Seed, Layers: 3, Wander: 0.12, MaxOff: r * 0.15, Filter: "none", Class: "nagare-led-dot"}, l.Theme)

	var anim strings.Builder
	for _, a := range l.Actions["blink"] {
		fmt.Fprintf(&anim, `<animate attributeName="opacity" begin="%v" dur="%v" values="%s;1;0.55;1;0.6;1;%s" keyTimes="0;0.15;0.3;0.45;0.6;0.85;1" repeatCount="1" fill="remove"/>`,
			a["begin"], a["dur"], d.InactiveOpacity, d.InactiveOpacity)
	}
	// The pulse recolours the dot: animate the pigment of the wash layers.
	var recolor strings.Builder
	for _, a := range l.Actions["blink"] {
		fmt.Fprintf(&recolor, `<animate attributeName="fill" begin="%v" dur="%v" values="%s" keyTimes="0;0.15;0.3;0.45;0.6;0.85;1" repeatCount="1" fill="remove"/>`,
			a["begin"], a["dur"], pulses)
	}
	// Wash puts the pigment on its inner group; the animation goes there.
	if i := strings.Index(wash, `<g transform=`); i >= 0 {
		if j := strings.Index(wash[i:], ">"); j >= 0 {
			k := i + j + 1
			wash = wash[:k] + recolor.String() + wash[k:]
		}
	}
	return fmt.Sprintf(`<g class="nagare-led" transform="translate(%s, %s)" opacity="%s">%s%s</g>`,
		style.F(l.X), style.F(l.Y), d.InactiveOpacity, anim.String(), wash)
}
