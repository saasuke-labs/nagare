package components

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/saasuke-labs/nagare/pkg/components"
)

type Controls struct {
	components.Shape
}

const ControlsTemplate = `<g transform="translate({{printf "%.6f" .ControlsX}},{{printf "%.6f" .ControlsY}})">
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="0" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#ff5f57"/>
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="{{printf "%.6f" .ControlSpacing}}" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#febc2e"/>
                        <circle r="{{printf "%.6f" .ControlRadius}}" cx="{{printf "%.6f" (mul .ControlSpacing 2)}}" cy="{{printf "%.6f" (mul .ControlRadius 1.33)}}" fill="#28c840"/>
                </g>`

type ControlsTemplateData struct {
	//X                      float64
	//Y                      float64
	//Width                  float64
	//Height                 float64
	CornerRadius float64
	//TopBarHeight           float64
	//UrlBarWidth            float64
	//UrlBarHeight           float64
	//UrlBarX                float64
	//UrlBarY                float64
	//ContentAreaWidth       float64
	//ContentAreaHeight      float64
	//ContentAreaX           float64
	//ContentAreaY           float64
	ControlsX      float64
	ControlsY      float64
	ControlRadius  float64
	ControlSpacing float64
	//FontSize               float64
	//BackgroundColor        string
	//ForegroundColor        string
	//	ContentBackgroundColor string
	//	URL                    string
	//	Text                   string
}

func (r *Controls) Draw() string {
	fmt.Println("Drawing Controls at", r.X, r.Y, "size", r.Width, r.Height)

	actualWidth := r.Width
	actualHeight := r.Height

	// // Calculate all dimensions
	// cornerRadius := actualWidth * 0.015625        // 10/640
	// topBarHeight := actualHeight * 0.1047619      // 44/420
	// urlBarHeight := actualHeight * 0.0571428      // 24/420
	// urlBarWidth := actualWidth * 0.75             // 480/640
	// urlBarX := actualWidth * 0.15625              // 100/640
	// urlBarY := actualHeight * 0.0238095           // 10/420
	// contentAreaWidth := actualWidth * 0.9625      // 616/640
	// contentAreaHeight := actualHeight * 0.8380952 // 352/420
	// contentAreaX := actualWidth * 0.01875         // 12/640
	// contentAreaY := actualHeight * 0.1333333      // 56/420
	controlsX := actualWidth * 0.021875      // 14/640
	controlsY := actualHeight * 0.0333333    // 14/420
	controlRadius := actualWidth * 0.009375  // 6/640
	controlSpacing := actualWidth * 0.028125 // 18/640
	// fontSize := actualWidth * 0.05                // 13/640

	// Create template data
	data := ControlsTemplateData{
		// X:                      r.X,
		// Y:                      r.Y,
		// Width:                  actualWidth,
		// Height:                 actualHeight,
		//CornerRadius: cornerRadius,
		// TopBarHeight:           topBarHeight,
		// UrlBarWidth:            urlBarWidth,
		// UrlBarHeight:           urlBarHeight,
		// UrlBarX:                urlBarX,
		// UrlBarY:                urlBarY,
		// ContentAreaWidth:       contentAreaWidth,
		// ContentAreaHeight:      contentAreaHeight,
		// ContentAreaX:           contentAreaX,
		// ContentAreaY:           contentAreaY,
		ControlsX:      controlsX,
		ControlsY:      controlsY,
		ControlRadius:  controlRadius,
		ControlSpacing: controlSpacing,
		// FontSize:               fontSize,
		// BackgroundColor:        r.Props.BackgroundColor,
		// ForegroundColor:        r.Props.ForegroundColor,
		// ContentBackgroundColor: r.Props.ContentBackgroundColor,
		// URL:                    r.Props.URL,
		// Text:                   r.Props.Text,
	}

	// Create and execute template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b float64) float64 { return a + b },
		"mul": func(a, b float64) float64 { return a * b },
	}

	tmpl := template.Must(template.New("Controls").Funcs(funcMap).Parse(ControlsTemplate))
	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return ""
	}
	return result.String()
}
