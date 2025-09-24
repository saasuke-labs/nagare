package renderer

import (
	"fmt"
	"nagare/layout"
)

// Render generates SVG code from a layout
func Render(l layout.Layout, canvasWidth, canvasHeight int) string {
	svg := fmt.Sprintf(`
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
	<!-- Background -->
	<rect width="%d" height="%d" fill="#ffffff"/>
	
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
	</text>
</svg>`,
		canvasWidth, canvasHeight,
		canvasWidth, canvasHeight,
		l.Bounds.X, l.Bounds.Y, l.Bounds.Width, l.Bounds.Height,
		l.Bounds.X+l.Bounds.Width/2, l.Bounds.Y+l.Bounds.Height/2,
		l.Text,
	)

	return svg
}
