package renderer

import (
	"fmt"
	"nagare/layout"
	"strings"
)

// renderElement generates SVG code for a single element
func renderElement(l layout.Layout) string {
	if len(l.Children) > 0 {
		// Container node
		return fmt.Sprintf(`
		<g transform="translate(%f,%f)">
			<!-- Container rectangle -->
			<rect 
				x="0" 
				y="0" 
				width="%f" 
				height="%f" 
				fill="#f0f0f0"
				stroke="#666666"
				stroke-width="2"/>
			
			<!-- Container Title -->
			<text 
				x="%f" 
				y="30"
				font-family="Arial" 
				font-size="16"
				font-weight="bold"
				fill="#333333"
				text-anchor="middle"
				dominant-baseline="middle">
				%s
			</text>
			<!-- Container content group -->
			<g transform="translate(16,60)">
				%s
			</g>
		</g>`,
			l.Bounds.X, l.Bounds.Y,
			l.Bounds.Width, l.Bounds.Height,
			l.Bounds.Width/2, // Center title
			l.Text,
			renderChildren(l.Children),
		)
	}

	// Leaf node
	return fmt.Sprintf(`
	<g transform="translate(%f,%f)">
		<!-- Element rectangle -->
		<rect 
			x="0" 
			y="0" 
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
		</text>
	</g>`,
		l.Bounds.X, l.Bounds.Y,
		l.Bounds.Width, l.Bounds.Height,
		l.Bounds.Width/2, l.Bounds.Height/2,
		l.Text,
	)
}

// renderChildren generates SVG code for child elements
func renderChildren(children []layout.Layout) string {
	var elements []string
	for _, child := range children {
		elements = append(elements, renderElement(child))
	}
	return strings.Join(elements, "\n")
}

// Render generates SVG code from a layout
func Render(l layout.Layout, canvasWidth, canvasHeight int) string {
	// Create the SVG wrapper with background and the layout
	return fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
	<!-- Background -->
	<rect width="%d" height="%d" fill="#ffffff"/>
	%s
</svg>`,
		canvasWidth, canvasHeight,
		canvasWidth, canvasHeight,
		renderElement(l),
	)
}
