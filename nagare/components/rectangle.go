package components

import "fmt"

type Component interface {
	Draw() string
}

type Shape struct {
	Width  float64
	Height float64
	X      float64
	Y      float64
}

type Rectangle struct {
	Shape
	Text string
}

func (r *Rectangle) Draw() string {
	fmt.Println("Drawing rectangle:", r.Text, "at", r.X, r.Y, "size", r.Width, r.Height)
	return fmt.Sprintf(`
        <g transform="translate(%f,%f)">
                <!-- Element rectangle -->
                <rect
                        x="0"
                        y="0"
                        width="%f"
                        height="%f"
                        fill="#333333"
                        stroke="#cccccc"
                        stroke-width="2"/>

                <!-- Text -->
                <text
                        x="%f"
                        y="%f"
                        font-family="Arial"
                        font-size="14"
                        fill="#cccccc"
                        text-anchor="middle"
                        dominant-baseline="middle">
                        %s
                </text>
        </g>`,
		r.X, r.Y,
		r.Width, r.Height,
		r.Width/2, r.Height/2,
		r.Text,
	)
}
