package layout

import (
	"fmt"
	"nagare/parser"
)

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

func (n Layout) String() string {
	if len(n.Children) == 0 {
		return n.Text
	}

	childrenStr := ""

	for _, child := range n.Children {
		childrenStr += "[ " + child.String() + " ]"
	}

	return fmt.Sprintf("%s [%s]", n.Text, childrenStr)

}

const (
	rectWidth       = 120.0
	rectHeight      = 60.0
	maxColumns      = 3
	mainPadding     = 16.0 // Padding for the main layout
	subPadding      = 8.0  // Padding for nested layouts
	titleHeight     = 40.0 // Height reserved for container title
	titleTopPadding = 20.0 // Padding above the container title
)

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout {
	if node.Type == parser.NODE_CONTAINER {
		// Container node
		childLayouts := make([]Layout, len(node.Children))

		// Calculate total width needed for children in a row
		childrenWidth := float64(len(node.Children)) * (rectWidth + mainPadding)
		containerWidth := max(childrenWidth, 260.0) // Min container width

		// Calculate height needed for container with title and children
		containerHeight := rectHeight + titleHeight + mainPadding*2

		// Position children in a row
		for i, child := range node.Children {
			childX := float64(i) * (rectWidth + mainPadding)
			childLayout := Calculate(child, rectWidth, rectHeight)
			childLayout.Bounds.X = childX
			childLayout.Bounds.Y = 0 // Y offset handled by group transform in renderer
			childLayouts[i] = childLayout
		}

		// Center container in its allocated space
		return Layout{
			Bounds: Rect{
				X:      (canvasWidth - containerWidth) / 2,
				Y:      mainPadding,
				Width:  containerWidth,
				Height: containerHeight,
			},
			Text:     node.Text,
			Children: childLayouts,
		}
	}

	if node.Text == "" {
		// Root node - arrange children horizontally with equal spacing
		childLayouts := make([]Layout, len(node.Children))
		spacing := float64(canvasWidth) / float64(len(node.Children)+1)

		for i, child := range node.Children {
			childX := spacing*float64(i+1) - rectWidth/2
			childLayout := Calculate(child, rectWidth*2, canvasHeight-mainPadding*2)
			childLayout.Bounds.X = childX
			childLayout.Bounds.Y = mainPadding
			childLayouts[i] = childLayout
		}

		return Layout{
			Bounds: Rect{
				X:      0,
				Y:      0,
				Width:  canvasWidth,
				Height: canvasHeight,
			},
			Children: childLayouts,
		}
	}

	// Regular node (leaf)
	return Layout{
		Bounds: Rect{
			X:      0,
			Y:      0,
			Width:  rectWidth,
			Height: rectHeight,
		},
		Text: node.Text,
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
