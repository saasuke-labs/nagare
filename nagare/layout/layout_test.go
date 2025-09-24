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
		expectedRows int
		expectedCols int
	}{
		{
			name: "single element",
			node: parser.Node{
				Type: parser.NODE_ELEMENT,
				Children: []parser.Node{
					{Type: parser.NODE_ELEMENT, Text: "Server"},
				},
			},
			canvasWidth:  400,
			canvasHeight: 300,
			expectedRows: 1,
			expectedCols: 1,
		},
		{
			name: "two elements",
			node: parser.Node{
				Type: parser.NODE_ELEMENT,
				Children: []parser.Node{
					{Type: parser.NODE_ELEMENT, Text: "Server1"},
					{Type: parser.NODE_ELEMENT, Text: "Server2"},
				},
			},
			canvasWidth:  400,
			canvasHeight: 300,
			expectedRows: 1,
			expectedCols: 2,
		},
		{
			name: "three elements",
			node: parser.Node{
				Type: parser.NODE_ELEMENT,
				Children: []parser.Node{
					{Type: parser.NODE_ELEMENT, Text: "Server1"},
					{Type: parser.NODE_ELEMENT, Text: "Server2"},
					{Type: parser.NODE_ELEMENT, Text: "Server3"},
				},
			},
			canvasWidth:  400,
			canvasHeight: 300,
			expectedRows: 1,
			expectedCols: 3,
		},
		{
			name: "four elements",
			node: parser.Node{
				Type: parser.NODE_ELEMENT,
				Children: []parser.Node{
					{Type: parser.NODE_ELEMENT, Text: "Server1"},
					{Type: parser.NODE_ELEMENT, Text: "Server2"},
					{Type: parser.NODE_ELEMENT, Text: "Server3"},
					{Type: parser.NODE_ELEMENT, Text: "Server4"},
				},
			},
			canvasWidth:  400,
			canvasHeight: 300,
			expectedRows: 2,
			expectedCols: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Calculate(tt.node, tt.canvasWidth, tt.canvasHeight)

			numChildren := len(got.Children)
			if numChildren != len(tt.node.Children) {
				t.Errorf("Calculate() returned %v children, want %v", numChildren, len(tt.node.Children))
			}

			// Verify grid layout
			if numChildren > 0 {
				// Get the first and last elements to check row/column distribution
				first := got.Children[0]
				// last := got.Children[numChildren-1]

				// Check number of unique Y positions (rows)
				yPositions := make(map[float64]bool)
				for _, child := range got.Children {
					yPositions[child.Bounds.Y] = true
				}
				if rows := len(yPositions); rows != tt.expectedRows {
					t.Errorf("Calculate() created %v rows, want %v", rows, tt.expectedRows)
				}

				// Check number of unique X positions per row (columns)
				xPositions := make(map[float64]bool)
				for _, child := range got.Children {
					if child.Bounds.Y == first.Bounds.Y { // Check first row
						xPositions[child.Bounds.X] = true
					}
				}
				if cols := len(xPositions); cols > tt.expectedCols {
					t.Errorf("Calculate() created %v columns in first row, want <= %v", cols, tt.expectedCols)
				}

				// Check that elements have same size
				for _, child := range got.Children {
					if child.Bounds.Width != rectWidth || child.Bounds.Height != rectHeight {
						t.Errorf("Calculate() child has size %vx%v, want %vx%v",
							child.Bounds.Width, child.Bounds.Height, rectWidth, rectHeight)
					}
				}

				// Check that elements preserve their text
				for i, child := range got.Children {
					if child.Text != tt.node.Children[i].Text {
						t.Errorf("Calculate() child %v has text %v, want %v",
							i, child.Text, tt.node.Children[i].Text)
					}
				}
			}
		})
	}
}
