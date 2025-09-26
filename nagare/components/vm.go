package components

import (
	"bytes"
	"fmt"
	"nagare/props"
	"text/template"
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
		ContentBackgroundColor: "#333333", // Dark content area by default
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
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

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

func (r *VM) Draw(colWidth, rowHeight float64) string {
	fmt.Println("Drawing VM at", r.X, r.Y, "size", r.Width, r.Height)

	actualWidth := float64(r.Width) * colWidth
	actualHeight := float64(r.Height) * rowHeight

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
		X:                      float64(r.X) * colWidth,
		Y:                      float64(r.Y) * rowHeight,
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

	// Calculate the usable area for children within the content area
	childAreaX := contentAreaX + actualWidth*0.02 // Add some padding from content area edge
	childAreaY := contentAreaY + actualHeight*0.02
	childAreaWidth := contentAreaWidth - actualWidth*0.04 // Subtract padding from both sides
	childAreaHeight := contentAreaHeight - actualHeight*0.04

	// If we have children, create a group for them with proper transformation
	if len(r.Children) > 0 {
		// Start the children group with translation to content area
		svg += fmt.Sprintf(`<g transform="translate(%f,%f)">`,
			float64(r.X)*colWidth+childAreaX,
			float64(r.Y)*rowHeight+childAreaY)

		// Calculate dimensions based on available space
		maxChildrenPerRow := 3                                                    // We can fit 3 servers side by side
		rowCount := (len(r.Children) + maxChildrenPerRow - 1) / maxChildrenPerRow // Round up division

		// Calculate child dimensions as portion of content area
		childWidth := childAreaWidth / float64(maxChildrenPerRow)
		childHeight := childAreaHeight / float64(rowCount)

		// Add some padding proportional to the smaller dimension
		padding := min(childWidth, childHeight) * 0.1
		effectiveChildWidth := childWidth - padding*2
		effectiveChildHeight := childHeight - padding*2

		// Draw each child
		for i, child := range r.Children {
			row := i / maxChildrenPerRow
			col := i % maxChildrenPerRow

			xPos := float64(col)*childWidth + padding
			yPos := float64(row)*childHeight + padding

			// Create a scaling transform for this child
			svg += fmt.Sprintf(`<g transform="translate(%f,%f)">`,
				xPos, yPos)

			// Update child's position and size
			if rect, ok := child.(interface{ SetBounds(x, y, w, h int) }); ok {
				rect.SetBounds(
					0, // Local coordinates
					0,
					int(effectiveChildWidth/colWidth), // Convert back to grid units
					int(effectiveChildHeight/rowHeight),
				)
			}

			// Draw the child with proper scaling
			childSVG := child.Draw(
				effectiveChildWidth/float64(int(effectiveChildWidth/colWidth)), // Scale to fit
				effectiveChildHeight/float64(int(effectiveChildHeight/rowHeight)),
			)
			svg += childSVG
			svg += "</g>"
		}

		// Close the children group
		svg += "</g>"
	}

	return svg
}
