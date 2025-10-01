package components

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/saasuke-labs/nagare/pkg/props"
)

const (
	VMContentAreaXRatio      = 0.01875
	VMContentAreaYRatio      = 0.1333333
	VMContentAreaWidthRatio  = 0.9625
	VMContentAreaHeightRatio = 0.8380952
)

// VMProps defines the configurable properties for a VM component
type VMProps struct {
	Title                  string `prop:"title"`
	BackgroundColor        string `prop:"bg"`
	ForegroundColor        string `prop:"fg"`
	ContentBackgroundColor string `prop:"contentBg"`
}

// Parse implements the Props interface
func (b *VMProps) Parse(input string) error {
	return props.ParseProps(input, b)
}

// DefaultVMProps returns a VMProps with default values
func DefaultVMProps() VMProps {
	return VMProps{
		Title:                  "",
		BackgroundColor:        "#e6f3ff",
		ForegroundColor:        "#333333",
		ContentBackgroundColor: "#ccc", // Light content area to keep connections visible
	}
}

type VM struct {
	Shape
	Text     string
	Props    VMProps
	State    string // Current state name
	Children []Component
}

// NewVM creates a new VM with default props
func NewVM() *VM {
	return &VM{
		Props:    DefaultVMProps(),
		Children: make([]Component, 0),
	}
}

// AddChild adds a child component to the VM
func (r *VM) AddChild(child Component) {
	r.Children = append(r.Children, child)
}

const VMTemplate = `<g transform="translate({{printf "%.6f" .X}},{{printf "%.6f" .Y}})">
                <g class="ns" filter="url(#softShadow)">
                        <rect x="0" y="0" width="{{printf "%.6f" .Width}}" height="{{printf "%.6f" .Height}}" rx="{{printf "%.6f" .CornerRadius}}" fill="{{.BackgroundColor}}" stroke="{{.ForegroundColor}}"/>
                        <rect x="0" y="0" width="{{printf "%.6f" .Width}}" height="{{printf "%.6f" .TopBarHeight}}" rx="{{printf "%.6f" .CornerRadius}}" ry="{{printf "%.6f" .CornerRadius}}" fill="{{.BackgroundColor}}" stroke="{{.ForegroundColor}}"/>
                        <rect x="{{printf "%.6f" .ContentAreaX}}" y="{{printf "%.6f" .ContentAreaY}}" width="{{printf "%.6f" .ContentAreaWidth}}" height="{{printf "%.6f" .ContentAreaHeight}}" rx="{{printf "%.6f" (mul .CornerRadius 0.6)}}" fill="{{.ContentBackgroundColor}}" stroke="{{.ForegroundColor}}" opacity="0.9"/>
                </g>
                <g transform="translate({{printf "%.6f" .ControlsX}},{{printf "%.6f" .ControlsY}})">
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="0" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#ff5f57"/>
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="{{printf "%.6f" .ControlSpacing}}" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#febc2e"/>
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="{{printf "%.6f" (mul .ControlSpacing 2)}}" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#28c840"/>
                        <!-- Title text positioned after controls -->
                        <text x="{{printf "%.6f" (add (mul .ControlSpacing 3) (mul .ControlRadius 2))}}" 
                              y="{{printf "%.6f" (mul .ControlRadius 1.33)}}" 
                              text-anchor="start" 
                              dominant-baseline="middle"
                              font-family="-apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
                              font-size="{{.FontSize}}" 
                              fill="{{.ForegroundColor}}">{{.Title}}</text>
                </g>
        </g>`

type VMTemplateData struct {
	X                      float64
	Y                      float64
	Width                  float64
	Height                 float64
	CornerRadius           float64
	TopBarHeight           float64
	ContentAreaWidth       float64
	ContentAreaHeight      float64
	ContentAreaX           float64
	ContentAreaY           float64
	ControlsX              float64
	ControlsY              float64
	ControlRadius          float64
	ControlSpacing         float64
	FontSize               float64
	BackgroundColor        string
	ForegroundColor        string
	ContentBackgroundColor string
	Title                  string
}

func (r *VM) Draw() string {
	fmt.Println("Drawing VM at", r.X, r.Y, "size", r.Width, r.Height)

	actualWidth := r.Width
	actualHeight := r.Height

	// Calculate all dimensions
	cornerRadius := actualWidth * 0.015625        // 10/640
	topBarHeight := actualHeight * 0.1047619      // 44/420
	contentAreaWidth := actualWidth * 0.9625      // 616/640
	contentAreaHeight := actualHeight * 0.8380952 // 352/420
	contentAreaX := actualWidth * 0.01875         // 12/640
	contentAreaY := actualHeight * 0.1333333      // 56/420
	controlsX := actualWidth * 0.021875           // 14/640
	controlsY := actualHeight * 0.0333333         // 14/420
	controlRadius := actualWidth * 0.009375       // 6/640
	controlSpacing := actualWidth * 0.028125      // 18/640
	fontSize := actualWidth * 0.05                // 13/640

	// Create template data
	data := VMTemplateData{
		X:                      r.X,
		Y:                      r.Y,
		Width:                  actualWidth,
		Height:                 actualHeight,
		CornerRadius:           cornerRadius,
		TopBarHeight:           topBarHeight,
		ContentAreaWidth:       contentAreaWidth,
		ContentAreaHeight:      contentAreaHeight,
		ContentAreaX:           contentAreaX,
		ContentAreaY:           contentAreaY,
		ControlsX:              controlsX,
		ControlsY:              controlsY,
		ControlRadius:          controlRadius,
		ControlSpacing:         controlSpacing,
		FontSize:               fontSize,
		BackgroundColor:        r.Props.BackgroundColor,
		ForegroundColor:        r.Props.ForegroundColor,
		ContentBackgroundColor: r.Props.ContentBackgroundColor,
		Title:                  r.Props.Title,
	}

	// Create and execute template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b float64) float64 { return a + b },
		"mul": func(a, b float64) float64 { return a * b },
	}

	tmpl := template.Must(template.New("VM").Funcs(funcMap).Parse(VMTemplate))
	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return ""
	}

	svg := result.String()

	// If we have children, position them relative to the VM's content area using
	// the provided bounds from the layout stage.
	if len(r.Children) > 0 {
		svg += fmt.Sprintf(`<g transform="translate(%f,%f)">`,
			r.X+contentAreaX,
			r.Y+contentAreaY)

		for _, child := range r.Children {
			svg += child.Draw()
		}

		svg += "</g>"
	}

	return svg
}
