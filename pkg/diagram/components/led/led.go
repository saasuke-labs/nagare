package led

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
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

type Legacy struct {
	components.Shape
	Text    string
	Props   Props
	Actions map[string][]map[string]any
}

func NewLegacy(id string) *Legacy {
	return &Legacy{Text: id, Props: DefaultProps(), Actions: map[string][]map[string]any{}}
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

func (l *Legacy) data() templateData {
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

func (l *Legacy) Draw() string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "led", l.data()); err != nil {
		return fmt.Sprintf("<!-- Error rendering led template: %v -->", err)
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
	comp.Actions = actionMapFromAny(nodeProps["actions"])
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
