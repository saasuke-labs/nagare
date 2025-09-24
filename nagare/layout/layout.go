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

const (
	rectWidth  = 120.0
	rectHeight = 60.0
	maxColumns = 3
)

// Calculate computes the layout for an AST
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout {
	if len(node.Children) == 0 {
		// Single node case - center it
		if node.Text != "" {
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
		return Layout{} // Empty layout for empty node
	}

	// Calculate grid dimensions
	numNodes := len(node.Children)
	numColumns := min(numNodes, maxColumns)
	numRows := (numNodes + maxColumns - 1) / maxColumns

	// Calculate the width and height of each grid cell
	cellWidth := canvasWidth / float64(numColumns)
	cellHeight := canvasHeight / float64(numRows)

	// Create layout for each child
	var children []Layout
	for i, child := range node.Children {
		row := i / maxColumns
		col := i % maxColumns

		// Calculate center position of the current grid cell
		cellCenterX := float64(col)*cellWidth + cellWidth/2
		cellCenterY := float64(row)*cellHeight + cellHeight/2

		// Position rectangle centered in the cell
		x := cellCenterX - rectWidth/2
		y := cellCenterY - rectHeight/2

		childLayout := Layout{
			Bounds: Rect{
				X:      x,
				Y:      y,
				Width:  rectWidth,
				Height: rectHeight,
			},
			Text: child.Text,
		}
		children = append(children, childLayout)
	}

	// Root layout encompasses all children
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
