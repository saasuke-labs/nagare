package components

import "fmt"

type Component interface {
	Draw(colWidth, rowHeight float64) string
}

type Shape struct {
	Width  int
	Height int
	X      int
	Y      int
}

type Rectangle struct {
	Shape
	Text string
}

func (r *Rectangle) Draw(colWidth, rowHeight float64) string {
	fmt.Println("Drawing rectangle:", r.Text, "at", r.X, r.Y, "size", r.Width, r.Height)
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
		float64(r.X)*colWidth, float64(r.Y)*rowHeight,
		float64(r.Width)*colWidth, float64(r.Height)*rowHeight,
		float64(r.Width/2)*colWidth, float64(r.Height/2)*rowHeight,
		r.Text,
	)
}
