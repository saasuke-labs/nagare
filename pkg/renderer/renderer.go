package renderer

import (
	"fmt"

	"github.com/saasuke-labs/nagare/pkg/components"
	"github.com/saasuke-labs/nagare/pkg/layout"
)

func renderChildren(children []components.Component) string {
	var elements string
	for _, child := range children {
		elements += child.Draw()
	}
	return elements
}

func drawGrid() string {
	return ""
}

// Render generates SVG code from a layout
func Render(l layout.Layout, canvasWidth, canvasHeight int) string {
	// Create the SVG wrapper with background and the layout
	return fmt.Sprintf(`<svg viewBox="0 0 %d %d" style="width:100%%; height:auto;" xmlns="http://www.w3.org/2000/svg">
        <!-- Background -->
        <rect width="%d" height="%d" fill="#ffffff"/>
        %s
        %s
</svg>`,
		canvasWidth, canvasHeight,
		canvasWidth, canvasHeight,
		drawGrid(),
		renderChildren(l.Children),
	)
}
