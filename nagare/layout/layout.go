package layout

import (
	"nagare/components"
	"nagare/parser"
)

const (
	DefaultColumns = 12
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
	Children []components.Component
}

// func (n Layout) String() string {
// 	if len(n.Children) == 0 {
// 		return n.Text
// 	}

// 	childrenStr := ""

// 	for _, child := range n.Children {
// 		childrenStr += "[ " + child.String() + " ]"
// 	}

// 	return fmt.Sprintf("%s [%s]", n.Text, childrenStr)

// }

// const (
// 	mainPadding     = 16.0 // Padding for the main layout
// 	subPadding      = 8.0  // Padding for nested layouts
// 	titleHeight     = 40.0 // Height reserved for container title
// 	titleTopPadding = 20.0 // Padding above the container title
// )

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout {
	// columns := 12
	// columnsWidth := float64(canvasWidth) / float64(columns)
	// rows := canvasHeight / columnsWidth
	// rowsHeight := canvasHeight / rows

	// rectWidth := 3 * columnsWidth
	// rectHeight := 2 * rowsHeight

	// if node.Type == parser.NODE_CONTAINER {
	// 	// Container node
	// 	childLayouts := make([]Layout, len(node.Children))

	// 	// Calculate total width needed for children in a row
	// 	childrenWidth := float64(len(node.Children)) * (rectWidth + mainPadding)
	// 	containerWidth := max(childrenWidth, 260.0) // Min container width

	// 	// Calculate height needed for container with title and children
	// 	containerHeight := rectHeight + titleHeight + mainPadding*2

	// 	// Position children in a row
	// 	for i, child := range node.Children {
	// 		childX := float64(i) * (rectWidth + mainPadding)
	// 		childLayout := Calculate(child, rectWidth, rectHeight)
	// 		childLayout.Bounds.X = childX
	// 		childLayout.Bounds.Y = 0 // Y offset handled by group transform in renderer
	// 		childLayouts[i] = childLayout
	// 	}

	// 	// Center container in its allocated space
	// 	return Layout{
	// 		Bounds: Rect{
	// 			X:      (canvasWidth - containerWidth) / 2,
	// 			Y:      mainPadding,
	// 			Width:  containerWidth,
	// 			Height: containerHeight,
	// 		},
	// 		Text:     node.Text,
	// 		Children: childLayouts,
	// 	}
	// }

	// if node.Text == "" {
	// 	// Root node - arrange children horizontally with equal spacing
	// 	childLayouts := make([]Layout, len(node.Children))
	// 	spacing := float64(canvasWidth) / float64(len(node.Children)+1)

	// 	for i, child := range node.Children {
	// 		childX := spacing*float64(i+1) - rectWidth/2
	// 		childLayout := Calculate(child, rectWidth*2, canvasHeight-mainPadding*2)
	// 		childLayout.Bounds.X = childX
	// 		childLayout.Bounds.Y = mainPadding
	// 		childLayouts[i] = childLayout
	// 	}

	// 	return Layout{
	// 		Bounds: Rect{
	// 			X:      0,
	// 			Y:      0,
	// 			Width:  canvasWidth,
	// 			Height: canvasHeight,
	// 		},
	// 		Children: childLayouts,
	// 	}
	// }

	// We start with the root node being a container.
	// Ignore it and go straight to its children.
	children := make([]components.Component, len(node.Children))
	for i, child := range node.Children {

		// FIME: HARDCODED TYPES
		if child.Type == "Browser" {
			children[i] = &components.Browser{
				Shape: components.Shape{
					Width:  3, // Based on Grid system. 3 cells x 2 cells
					Height: 3,
					X:      int(float64(i*4) + 1), // 3 cells width + 1 cell gap
					Y:      0,
				},
				Text: child.Text,
			}
			continue
		} else {
			children[i] = &components.Rectangle{
				Shape: components.Shape{
					Width:  3, // Based on Grid system. 3 cells x 2 cells
					Height: 2,
					X:      int(float64(i*4) + 1), // 3 cells width + 1 cell gap
					Y:      0,
				},
				Text: child.Text,
			}
		}
	}

	return Layout{
		Bounds: Rect{
			X:      0,
			Y:      0,
			Width:  canvasWidth,
			Height: canvasHeight,
		},
		Children: children,
	}

}

// func max(a, b float64) float64 {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// func minf(a, b float64) float64 {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }
