package renderer

import (
	"fmt"
	"nagare/layout"
	"strings"
)

// renderElement generates SVG code for a single element
func renderElement(l layout.Layout) string {
	return fmt.Sprintf(`
	<!-- Element rectangle -->
	<rect 
		x="%f" 
		y="%f" 
		width="%f" 
		height="%f" 
		fill="#cccccc"
		stroke="#333333"
		stroke-width="2"/>
	
	<!-- Text -->
	<text 
		x="%f" 
		y="%f" 
		font-family="Arial" 
		font-size="14"
		fill="#333333"
		text-anchor="middle"
		dominant-baseline="middle">
		%s
	</text>`,
		l.Bounds.X, l.Bounds.Y, l.Bounds.Width, l.Bounds.Height,
		l.Bounds.X+l.Bounds.Width/2, l.Bounds.Y+l.Bounds.Height/2,
		l.Text,
	)
}

// Render generates SVG code from a layout
func Render(l layout.Layout, canvasWidth, canvasHeight int) string {
	var elements []string

	// If it's a single node with text
	if len(l.Children) == 0 && l.Text != "" {
		elements = append(elements, renderElement(l))
	}

	// Add all child elements
	for _, child := range l.Children {
		elements = append(elements, renderElement(child))
	}

	svg := fmt.Sprintf(`
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
	<!-- Background -->
	<rect width="%d" height="%d" fill="#ffffff"/>
	%s
</svg>`,
		canvasWidth, canvasHeight,
		canvasWidth, canvasHeight,
		strings.Join(elements, "\n"),
	)

	return svg
}
