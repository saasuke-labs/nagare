package components

import (
	"fmt"
	"html/template"
	"math"
	"strings"

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
	TitleBarX              float64
	TitleBarY              float64
	FontSize               float64
	BackgroundColor        string
	ForegroundColor        string
	ContentBackgroundColor string
	Title                  string
	HeaderControlProps     HeaderControlProps
	ChildrenContent        template.HTML
}

func (r *VM) Draw() string {
	fmt.Println("Drawing VM at", r.X, r.Y, "size", r.Width, r.Height)

	actualWidth := r.Width
	actualHeight := r.Height

	// Calculate all dimensions
	cornerRadius := actualWidth * 0.015625                 // 10/640
	topBarHeight := actualHeight * 0.1047619               // 44/420
	contentAreaWidth := actualWidth * 0.9625               // 616/640
	contentAreaHeight := actualHeight * 0.8380952          // 352/420
	contentAreaX := actualWidth * 0.01875                  // 12/640
	contentAreaY := actualHeight * 0.1333333               // 56/420
	fontSize := math.Min(actualHeight, actualWidth) * 0.07 // 13/640
	titleBarX := actualWidth * 0.15625                     // 100/640
	titleBarY := actualHeight * 0.0538095                  // 10/420

	// Render children within the content area
	var childrenContent strings.Builder
	if len(r.Children) > 0 {
		// Create a group for all children with proper translation to content area
		childrenContent.WriteString(fmt.Sprintf(`<g transform="translate(%f,%f)" class="vm-content">`,
			contentAreaX,
			contentAreaY))

		for _, child := range r.Children {
			// Each child is rendered within the VM's content area
			childrenContent.WriteString(child.Draw())
		}

		childrenContent.WriteString("</g>")
	}

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
		TitleBarX:              titleBarX,
		TitleBarY:              titleBarY,
		FontSize:               fontSize,
		BackgroundColor:        r.Props.BackgroundColor,
		ForegroundColor:        r.Props.ForegroundColor,
		ContentBackgroundColor: r.Props.ContentBackgroundColor,
		Title:                  r.Props.Title,
		HeaderControlProps:     NewHeaderControlProps(actualWidth, actualHeight),
		ChildrenContent:        template.HTML(childrenContent.String()),
	}

	result, err := RenderTemplate("vm", data)

	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		return ""
	}

	return result
}
