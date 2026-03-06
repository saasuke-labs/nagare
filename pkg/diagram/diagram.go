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
	svg, _, _, err := RenderToSVGWithSize(code)
	return svg, err
}

// CreateDiagramWithSize generates an SVG diagram and returns the SVG along with the computed canvas size.
func CreateDiagramWithSize(code string) (string, int, int, error) {
	return RenderToSVGWithSize(code)
}

// RenderToSVG renders diagram DSL input into SVG.
func RenderToSVG(code string) (string, error) {
	svg, _, _, err := RenderToSVGWithSize(code)
	return svg, err
}

// RenderToSVGWithSize renders a diagram and returns SVG with computed canvas size.
func RenderToSVGWithSize(code string) (string, int, int, error) {
	tokens := tokenizer.Tokenize(code)
	ast, err := parser.Parse(tokens)
	if err != nil {
		return "", 0, 0, fmt.Errorf("parse error: %w", err)
	}

	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

	canvasWidth := normalizedCanvasDimension(l.Bounds.Width, defaultCanvasWidth)
	canvasHeight := normalizedCanvasDimension(l.Bounds.Height, defaultCanvasHeight)
	svg := renderer.Render(l, canvasWidth, canvasHeight)

	return svg, canvasWidth, canvasHeight, nil
}

// RenderToSVGWithDebug renders a diagram and includes parser/layout debug output.
func RenderToSVGWithDebug(code string) (string, error) {
	fmt.Printf("Input code:\n%s\n", code)

	tokens := tokenizer.Tokenize(code)
	fmt.Printf("Tokens: %+v\n", tokens)

	ast, err := parser.Parse(tokens)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("AST: \n%+v\n", ast)

	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)
	fmt.Printf("Layout: \n%+v\n", l)

	canvasWidth := normalizedCanvasDimension(l.Bounds.Width, defaultCanvasWidth)
	canvasHeight := normalizedCanvasDimension(l.Bounds.Height, defaultCanvasHeight)
	svg := renderer.Render(l, canvasWidth, canvasHeight)
	fmt.Println(svg)

	return svg, nil
}

func normalizedCanvasDimension(calculated, fallback float64) int {
	dim := int(calculated)
	if dim == 0 {
		dim = int(fallback)
	}
	return dim
}
