package renderer

import (
	"fmt"
	"nagare/layout"
)

// Render generates HTML canvas code from a layout
func Render(l layout.Layout, canvasWidth, canvasHeight int) string {
	html := fmt.Sprintf(`
<canvas id="diagram" width="%d" height="%d"></canvas>
<script>
const canvas = document.getElementById('diagram');
const ctx = canvas.getContext('2d');

// Set white background
ctx.fillStyle = '#ffffff';
ctx.fillRect(0, 0, %d, %d);

// Draw rectangle
ctx.fillStyle = '#cccccc';
ctx.strokeStyle = '#333333';
ctx.lineWidth = 2;
ctx.fillRect(%f, %f, %f, %f);
ctx.strokeRect(%f, %f, %f, %f);

// Draw text
ctx.fillStyle = '#333333';
ctx.font = '14px Arial';
ctx.textAlign = 'center';
ctx.textBaseline = 'middle';
ctx.fillText('%s', %f, %f);
</script>`,
		canvasWidth, canvasHeight,
		canvasWidth, canvasHeight,
		l.Bounds.X, l.Bounds.Y, l.Bounds.Width, l.Bounds.Height,
		l.Bounds.X, l.Bounds.Y, l.Bounds.Width, l.Bounds.Height,
		l.Text,
		l.Bounds.X+l.Bounds.Width/2, l.Bounds.Y+l.Bounds.Height/2,
	)

	return html
}
