package browser

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
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

//go:embed *.html
var templateFiles embed.FS

var tmpl = template.Must(
	template.New("").Funcs(template.FuncMap{
		"add": func(a, b float64) float64 { return a + b },
		"mul": func(a, b float64) float64 { return a * b },
		"sub": func(a, b float64) float64 { return a - b },
	}).ParseFS(templateFiles, "*.html"),
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
	return Props{
		BackgroundColor:        "#e6f3ff",
		ForegroundColor:        "#333333",
		ContentBackgroundColor: "#ffffff",
	}
}

// Component represents a browser diagram component.
type Component struct {
	components.Shape
	Text    string
	Props   Props
	State   string
	Actions map[string][]map[string]any
}

// New creates a new Component with default props.
func New(id string) *Component {
	return &Component{
		Text:    id,
		Props:   DefaultProps(),
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

// Draw renders the browser as SVG by composing its atom sub-components.
func (c *Component) Draw() string {
	return drawComposite(c.Shape, c.Props)
}

// DrawFromRenderNode is the entry point called by the diagram rendering pipeline.
func DrawFromRenderNode(id string, nodeProps map[string]any) string {
	comp := New(id)
	comp.Shape = core.ShapeFromProps(nodeProps, DefaultWidth, DefaultHeight)
	if raw, ok := nodeProps["_rawProps"].(string); ok {
		_ = comp.Props.Parse(raw)
	}
	comp.Actions = core.ActionMapFromAny(nodeProps["actions"])
	return comp.Draw()
}

// dimensions holds all pre-computed geometry values for rendering.
type dimensions struct {
	Width                  float64
	Height                 float64
	CornerRadius           float64
	TopBarHeight           float64
	UrlBarWidth            float64
	UrlBarHeight           float64
	UrlBarX                float64
	UrlBarY                float64
	ContentAreaWidth       float64
	ContentAreaHeight      float64
	ContentAreaX           float64
	ContentAreaY           float64
	FontSize               float64
	ControlsX              float64
	ControlsY              float64
	ControlRadius          float64
	ControlSpacing         float64
	BackgroundColor        string
	ForegroundColor        string
	ContentBackgroundColor string
	URL                    string
	Text                   string
}

// computeDimensions derives all proportional geometry values from the shape and props.
func computeDimensions(shape components.Shape, p Props) dimensions {
	w, h := shape.Width, shape.Height
	return dimensions{
		Width:                  w,
		Height:                 h,
		CornerRadius:           w * 0.015625,  // 10/640
		TopBarHeight:           h * 0.1047619, // 44/420
		UrlBarHeight:           h * 0.0571428, // 24/420
		UrlBarWidth:            w * 0.75,      // 480/640
		UrlBarX:                w * 0.15625,   // 100/640
		UrlBarY:                h * 0.0238095, // 10/420
		ContentAreaWidth:       w * 0.9625,    // 616/640
		ContentAreaHeight:      h * 0.8380952, // 352/420
		ContentAreaX:           w * 0.01875,   // 12/640
		ContentAreaY:           h * 0.1333333, // 56/420
		FontSize:               w * 0.05,      // ~32/640
		ControlsX:              w * 0.021875,  // 14/640
		ControlsY:              h * 0.0238095, // 10/420
		ControlRadius:          w * 0.009375,  // 6/640
		ControlSpacing:         w * 0.028125,  // 18/640
		BackgroundColor:        p.BackgroundColor,
		ForegroundColor:        p.ForegroundColor,
		ContentBackgroundColor: p.ContentBackgroundColor,
		URL:                    p.URL,
		Text:                   p.Text,
	}
}

// drawComposite builds the browser SVG by composing its three atom sub-components:
// application (outer chrome + content area + control buttons),
// titlebar (URL bar), and html-content (page label).
func drawComposite(shape components.Shape, p Props) string {
	dims := computeDimensions(shape, p)
	return renderSubTemplate("browser-application", dims) +
		renderSubTemplate("browser-titlebar", dims) +
		renderSubTemplate("browser-html-content", dims)
}

func renderSubTemplate(name string, data any) string {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return fmt.Sprintf("<!-- Error rendering %s template: %v -->", name, err)
	}
	return buf.String()
}
