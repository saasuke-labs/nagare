package diagram

import (
	"fmt"

	"github.com/saasuke-labs/nagare/pkg/layout"
	"github.com/saasuke-labs/nagare/pkg/parser"
	"github.com/saasuke-labs/nagare/pkg/renderer"
	"github.com/saasuke-labs/nagare/pkg/tokenizer"
)

// CreateDiagram generates an SVG diagram from the provided code and returns it as a string.
func CreateDiagram(code string) (string, error) {
	svg, _, _, err := CreateDiagramWithSize(code)
	return svg, err
}

// CreateDiagramWithSize generates an SVG diagram and returns the SVG along with the computed canvas size.
func CreateDiagramWithSize(code string) (string, int, int, error) {
	fmt.Printf("Input code:\n%s\n", string(code))

	// Pipeline:
	// 1. Tokenize
	tokens := tokenizer.Tokenize(string(code))
	fmt.Printf("Tokens: %+v\n", tokens)

	// 2. Parse
	ast, err := parser.Parse(tokens)
	if err != nil {
		return "", 0, 0, fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("AST: \n%+v\n", ast)

	// 3. Layout
	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

	fmt.Printf("Layout: \n%+v\n", l)

	// 4. Render using the computed layout dimensions
	canvasWidth := int(l.Bounds.Width)
	canvasHeight := int(l.Bounds.Height)
	if canvasWidth == 0 {
		canvasWidth = int(defaultCanvasWidth)
	}
	if canvasHeight == 0 {
		canvasHeight = int(defaultCanvasHeight)
	}

	html := renderer.Render(l, canvasWidth, canvasHeight)
	fmt.Println(html)
	return html, canvasWidth, canvasHeight, nil
}
