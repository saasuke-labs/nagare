package browser

import (
	"math"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	DefaultWidth  = 640.0
	DefaultHeight = 420.0

	// AllowedActions lists the action names the Browser component supports.
	// "request" generates a solid-line arrow to the target.
	// "response" generates a dashed-line arrow to the target.
	AllowedActions = "request,response"
)

// Props holds configurable properties for a Browser component.
type Props struct {
	URL                    string `prop:"url"`
	BackgroundColor        string `prop:"bg"`
	ForegroundColor        string `prop:"fg"`
	ContentBackgroundColor string `prop:"contentBg"`
	Text                   string `prop:"text"`
}

// Parse implements the propertyParser interface used by the layout package.
func (p *Props) Parse(input string) error {
	return props.ParseProps(input, p)
}

// DefaultProps returns a Props with default values.
func DefaultProps() Props {
	return Props{}
}

// Component represents a browser diagram component.
type Component struct {
	components.Shape
	Text    string
	Props   Props
	State   string
	Actions map[string][]map[string]any
	Theme   style.Theme
}

// New creates a new Component with default props.
func New(id string) *Component {
	return &Component{
		Text:    id,
		Props:   DefaultProps(),
		Theme:   style.Light,
		Actions: make(map[string][]map[string]any),
	}
}

// SetShape sets the component's geometry.
func (c *Component) SetShape(shape components.Shape) { c.Shape = shape }

// Offset adjusts the component's position by subtracting the given offsets.
func (c *Component) Offset(dx, dy float64) {
	c.X -= dx
	c.Y -= dy
}

// Draw renders the browser as SVG.
func (c *Component) Draw() string {
	return drawWindow(c.Shape, c.Props, c.Theme)
}

// ResolveAction implements core.ActionConnector. It maps Browser action names
// to ConnectionIntents so that layout can create arrows without knowing the
// per-action semantics:
//   - "request"  → solid-line arrow to the target
//   - "response" → dashed-line arrow to the target
//
// actionProps are the parsed key/value pairs from the action call, e.g.
// target:@s,dir:lr.
func (c *Component) ResolveAction(sourceID, actionName string, actionProps map[string]any) []core.ConnectionIntent {
	if actionName != "request" && actionName != "response" {
		return nil
	}
	targetRaw, _ := actionProps["target"].(string)
	if targetRaw == "" {
		return nil
	}
	targetID := strings.TrimPrefix(targetRaw, "@")
	dir, _ := actionProps["dir"].(string)
	style := ""
	if actionName == "response" {
		style = "dashed"
	}
	return []core.ConnectionIntent{{
		FromID: sourceID,
		ToID:   targetID,
		Dir:    dir,
		Style:  style,
	}}
}

// DrawFromRenderNode is the entry point called by the diagram rendering pipeline.
func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := New(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	comp.Theme = style.ThemeFromProps(nodeProps)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	comp.Actions = core.ActionMapFromAny(nodeProps["actions"])
	return comp.Draw()
}

// Content-area ratios, shared with the VM so children lay out identically.
const (
	contentXRatio = 0.01875
	contentYRatio = 0.1333333
	contentWRatio = 0.9625
	contentHRatio = 0.8380952
)

// drawWindow renders the browser: shared window chrome with the URL as an
// address pill, and an optional label centred on the page.
func drawWindow(shape components.Shape, p Props, t style.Theme) string {
	w, h := shape.Width, shape.Height
	cx, cy, cw, ch := w*contentXRatio, h*contentYRatio, w*contentWRatio, h*contentHRatio

	contentFill := style.Or(p.ContentBackgroundColor, t.Surface)

	var b strings.Builder
	b.WriteString(style.Window(style.WindowSpec{
		X: shape.X, Y: shape.Y, W: w, H: h,
		Class: "nagare-browser",
		BarH:  cy,
		CX:    cx, CY: cy, CW: cw, CH: ch,
		Title: p.URL, TitlePill: true,
		Bg: p.BackgroundColor, Fg: p.ForegroundColor,
		ContentFill: contentFill,
		Theme:       t,
	}))
	if p.Text != "" {
		size := style.Clamp(math.Min(cw, ch)*0.09, 10, 22)
		b.WriteString(style.Text(p.Text, style.TextOpts{
			X: cx + cw/2, Y: cy + ch/2, Size: size,
			Color: style.Legible(p.ForegroundColor, contentFill, t.Muted), Anchor: "middle", Width: cw * 0.9,
		}))
	}
	b.WriteString(style.WindowEnd)
	return b.String()
}
