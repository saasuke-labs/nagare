package layout

import "nagare/parser"

// Rect represents a rectangle in the layout
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Layout represents the computed layout of an element
type Layout struct {
	Bounds   Rect
	Text     string
	Children []Layout
}

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout {
	// For now, we create a fixed-size rectangle in the center
	const (
		rectWidth  = 120
		rectHeight = 60
	)

	return Layout{
		Bounds: Rect{
			X:      (canvasWidth - rectWidth) / 2,
			Y:      (canvasHeight - rectHeight) / 2,
			Width:  rectWidth,
			Height: rectHeight,
		},
		Text: node.Text,
	}
}
