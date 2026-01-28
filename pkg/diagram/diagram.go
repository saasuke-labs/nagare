package diagram

import (
	"github.com/saasuke-labs/nagare/pkg/layout"
	"github.com/saasuke-labs/nagare/pkg/nagare"
	"github.com/saasuke-labs/nagare/pkg/parser"
	"github.com/saasuke-labs/nagare/pkg/tokenizer"
)

// CreateDiagram generates an SVG diagram from the provided code and returns it as a string.
func CreateDiagram(code string) (string, error) {
	svg, _, _, err := CreateDiagramWithSize(code)
	return svg, err
}

// CreateDiagramWithSize generates an SVG diagram and returns the SVG along with the computed canvas size.
func CreateDiagramWithSize(code string) (string, int, int, error) {
	// Render the SVG
	svg, err := nagare.RenderToSVG(code)
	if err != nil {
		return "", 0, 0, err
	}

	// Calculate layout to get dimensions
	tokens := tokenizer.Tokenize(code)
	ast, parseErr := parser.Parse(tokens)
	if parseErr != nil {
		// Should not happen since nagare.RenderToSVG succeeded
		return svg, 800, 400, nil
	}

	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

	canvasWidth := int(l.Bounds.Width)
	canvasHeight := int(l.Bounds.Height)
	if canvasWidth == 0 {
		canvasWidth = int(defaultCanvasWidth)
	}
	if canvasHeight == 0 {
		canvasHeight = int(defaultCanvasHeight)
	}

	return svg, canvasWidth, canvasHeight, nil
}
