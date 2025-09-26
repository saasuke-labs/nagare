package components

import (
	"bytes"
	"fmt"
	"nagare/props"
	"text/template"
)

// VMProps defines the configurable properties for a VM component
type VMProps struct {
	URL             string `prop:"url"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	Text            string `prop:"text"`
}

// Parse implements the Props interface
func (b *VMProps) Parse(input string) error {
	return props.ParseProps(input, b)
}

// DefaultVMProps returns a VMProps with default values
func DefaultVMProps() VMProps {
	return VMProps{
		URL:             "",
		BackgroundColor: "#e6f3ff",
		ForegroundColor: "#333333",
		Text:            "",
	}
}

type VM struct {
	Shape
	Text  string
	Props VMProps
	State string // Current state name
}

// NewVM creates a new VM with default props
func NewVM() *VM {
	return &VM{
		Props: DefaultVMProps(),
	}
}

const VMTemplate = `<g transform="translate({{printf "%.6f" .X}},{{printf "%.6f" .Y}})">
                <g class="ns" filter="url(#softShadow)">
                        <rect x="0" y="0" width="{{printf "%.6f" .Width}}" height="{{printf "%.6f" .Height}}" rx="{{printf "%.6f" .CornerRadius}}" fill="{{.BackgroundColor}}" stroke="{{.ForegroundColor}}"/>
                        <rect x="0" y="0" width="{{printf "%.6f" .Width}}" height="{{printf "%.6f" .TopBarHeight}}" rx="{{printf "%.6f" .CornerRadius}}" ry="{{printf "%.6f" .CornerRadius}}" fill="{{.BackgroundColor}}" stroke="{{.ForegroundColor}}"/>
                        <rect x="{{printf "%.6f" .UrlBarX}}" y="{{printf "%.6f" .UrlBarY}}" width="{{printf "%.6f" .UrlBarWidth}}" height="{{printf "%.6f" .UrlBarHeight}}" rx="{{printf "%.6f" (mul .CornerRadius 0.6)}}" fill="#fff" stroke="{{.ForegroundColor}}" opacity="0.85"/>
                        <rect x="{{printf "%.6f" .ContentAreaX}}" y="{{printf "%.6f" .ContentAreaY}}" width="{{printf "%.6f" .ContentAreaWidth}}" height="{{printf "%.6f" .ContentAreaHeight}}" rx="{{printf "%.6f" (mul .CornerRadius 0.6)}}" fill="{{.BackgroundColor}}" stroke="{{.ForegroundColor}}" opacity="0.9"/>
                </g>
                <text x="{{printf "%.6f" (add .UrlBarX (mul .UrlBarWidth 0.5))}}" y="{{printf "%.6f" (add .UrlBarY (mul .UrlBarHeight 0.5))}}" text-anchor="middle" dominant-baseline="middle"
                        font-family="-apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
                        font-size="{{printf "%.6f" (mul .FontSize 0.8)}}" fill="{{.ForegroundColor}}">{{.URL}}</text>
                <text x="{{printf "%.6f" (mul .Width 0.5)}}" y="{{printf "%.6f" (add .ContentAreaY (mul .ContentAreaHeight 0.5))}}" text-anchor="middle" dominant-baseline="middle"
                        font-family="-apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif"
                        font-size="{{mul .FontSize 2}}" fill="{{.ForegroundColor}}">{{.Text}}</text>
                <g transform="translate({{printf "%.6f" .ControlsX}},{{printf "%.6f" .ControlsY}})">
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="0" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#ff5f57"/>
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="{{printf "%.6f" .ControlSpacing}}" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#febc2e"/>
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="{{printf "%.6f" (mul .ControlSpacing 2)}}" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#28c840"/>
                </g>
        </g>`

type VMTemplateData struct {
	X                 float64
	Y                 float64
	Width             float64
	Height            float64
	CornerRadius      float64
	TopBarHeight      float64
	UrlBarWidth       float64
	UrlBarHeight      float64
	UrlBarX           float64
	UrlBarY           float64
	ContentAreaWidth  float64
	ContentAreaHeight float64
	ContentAreaX      float64
	ContentAreaY      float64
	ControlsX         float64
	ControlsY         float64
	ControlRadius     float64
	ControlSpacing    float64
	FontSize          float64
	BackgroundColor   string
	ForegroundColor   string
	URL               string
	Text              string
}

func (r *VM) Draw(colWidth, rowHeight float64) string {
	fmt.Println("Drawing VM at", r.X, r.Y, "size", r.Width, r.Height)

	actualWidth := float64(r.Width) * colWidth
	actualHeight := float64(r.Height) * rowHeight

	// Calculate all dimensions
	cornerRadius := actualWidth * 0.015625        // 10/640
	topBarHeight := actualHeight * 0.1047619      // 44/420
	urlBarHeight := actualHeight * 0.0571428      // 24/420
	urlBarWidth := actualWidth * 0.75             // 480/640
	urlBarX := actualWidth * 0.15625              // 100/640
	urlBarY := actualHeight * 0.0238095           // 10/420
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
		X:                 float64(r.X) * colWidth,
		Y:                 float64(r.Y) * rowHeight,
		Width:             actualWidth,
		Height:            actualHeight,
		CornerRadius:      cornerRadius,
		TopBarHeight:      topBarHeight,
		UrlBarWidth:       urlBarWidth,
		UrlBarHeight:      urlBarHeight,
		UrlBarX:           urlBarX,
		UrlBarY:           urlBarY,
		ContentAreaWidth:  contentAreaWidth,
		ContentAreaHeight: contentAreaHeight,
		ContentAreaX:      contentAreaX,
		ContentAreaY:      contentAreaY,
		ControlsX:         controlsX,
		ControlsY:         controlsY,
		ControlRadius:     controlRadius,
		ControlSpacing:    controlSpacing,
		FontSize:          fontSize,
		BackgroundColor:   r.Props.BackgroundColor,
		ForegroundColor:   r.Props.ForegroundColor,
		URL:               r.Props.URL,
		Text:              r.Props.Text,
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
	return result.String()
}
