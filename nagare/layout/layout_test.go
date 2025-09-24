package layout

import (
	"nagare/parser"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name         string
		node         parser.Node
		canvasWidth  float64
		canvasHeight float64
	}{
		{
			name: "center element",
			node: parser.Node{
				Type: parser.NODE_ELEMENT,
				Text: "Server",
			},
			canvasWidth:  400,
			canvasHeight: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.node, tt.canvasWidth, tt.canvasHeight)

			// Test that the element is centered
			expectedX := (tt.canvasWidth - got.Bounds.Width) / 2
			expectedY := (tt.canvasHeight - got.Bounds.Height) / 2

			if got.Bounds.X != expectedX {
				t.Errorf("Calculate().Bounds.X = %v, want %v", got.Bounds.X, expectedX)
			}
			if got.Bounds.Y != expectedY {
				t.Errorf("Calculate().Bounds.Y = %v, want %v", got.Bounds.Y, expectedY)
			}

			// Test that the text is preserved
			if got.Text != tt.node.Text {
				t.Errorf("Calculate().Text = %v, want %v", got.Text, tt.node.Text)
			}

			// Test that the size is reasonable
			if got.Bounds.Width <= 0 || got.Bounds.Height <= 0 {
				t.Errorf("Calculate() returned invalid bounds size: %v x %v", got.Bounds.Width, got.Bounds.Height)
			}
		})
	}
}
