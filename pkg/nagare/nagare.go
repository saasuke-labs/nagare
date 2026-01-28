package nagare

import (
	"fmt"
	"os"

	"github.com/saasuke-labs/nagare/pkg/layout"
	"github.com/saasuke-labs/nagare/pkg/parser"
	"github.com/saasuke-labs/nagare/pkg/renderer"
	"github.com/saasuke-labs/nagare/pkg/tokenizer"
)

// RenderToSVG takes nagare DSL code as a string and returns the rendered SVG as a string.
// This is the main entry point for using nagare as a library.
func RenderToSVG(code string) (string, error) {
	// Pipeline:
	// 1. Tokenize
	tokens := tokenizer.Tokenize(code)

	// 2. Parse
	ast, err := parser.Parse(tokens)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	// 3. Layout
	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

	// 4. Render using the computed layout dimensions
	canvasWidth := int(l.Bounds.Width)
	canvasHeight := int(l.Bounds.Height)
	if canvasWidth == 0 {
		canvasWidth = int(defaultCanvasWidth)
	}
	if canvasHeight == 0 {
		canvasHeight = int(defaultCanvasHeight)
	}

	svg := renderer.Render(l, canvasWidth, canvasHeight)
	return svg, nil
}

// RenderToSVGWithDebug is like RenderToSVG but prints debug information to stdout.
func RenderToSVGWithDebug(code string) (string, error) {
	fmt.Printf("Input code:\n%s\n", code)

	// Pipeline:
	// 1. Tokenize
	tokens := tokenizer.Tokenize(code)
	fmt.Printf("Tokens: %+v\n", tokens)

	// 2. Parse
	ast, err := parser.Parse(tokens)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
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

	svg := renderer.Render(l, canvasWidth, canvasHeight)
	fmt.Println(svg)
	return svg, nil
}

// RenderFileToFile reads a nagare file from inputPath and writes the SVG output to outputPath.
func RenderFileToFile(inputPath, outputPath string) error {
	// Read input file
	code, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Render to SVG
	svg, err := RenderToSVG(string(code))
	if err != nil {
		return fmt.Errorf("failed to render diagram: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(svg), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
