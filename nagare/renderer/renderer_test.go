package renderer

import (
	"nagare/layout"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name         string
		layout       layout.Layout
		canvasWidth  int
		canvasHeight int
		checks       []string
	}{
		{
			name: "basic render",
			layout: layout.Layout{
				Bounds: layout.Rect{
					X:      150,
					Y:      120,
					Width:  100,
					Height: 60,
				},
				Text: "Server",
			},
			canvasWidth:  400,
			canvasHeight: 300,
			checks: []string{
				`<svg width="400" height="300"`, // SVG dimensions
				`fill="#ffffff"`,                // Background color
				`fill="#cccccc"`,                // Rectangle fill
				`stroke="#333333"`,              // Rectangle stroke
				`stroke-width="2"`,              // Border width
				`text-anchor="middle"`,          // Text centering
				`dominant-baseline="middle"`,    // Vertical text centering
				`Server`,                        // The actual text
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.layout, tt.canvasWidth, tt.canvasHeight)

			for _, check := range tt.checks {
				if !strings.Contains(got, check) {
					t.Errorf("Render() missing expected content: %q", check)
				}
			}
		})
	}
}
