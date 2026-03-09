package nagare

import (
	"fmt"
	"os"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/chart"
	"github.com/saasuke-labs/nagare/pkg/diagram"
)

// RenderToSVG takes nagare DSL code as a string and returns the rendered SVG as a string.
// This is the main entry point for using nagare as a library.
// Automatically detects whether the input is a chart or diagram.
func RenderToSVG(code string) (string, error) {
	input := strings.TrimSpace(code)

	// Check if this is a chart definition
	if strings.HasPrefix(input, "chart") {
		c, err := chart.Parse(input)
		if err != nil {
			return "", fmt.Errorf("chart parse error: %w", err)
		}
		return c.RenderSVG(), nil
	}

	return diagram.CreateDiagram(code)
}

// RenderToHTML is the legacy endpoint helper used by the HTTP server.
// For diagrams it returns the SVG; for charts it returns the full HTML page.
func RenderToHTML(code string) (string, error) {
	input := strings.TrimSpace(code)

	if strings.HasPrefix(input, "chart") {
		c, err := chart.Parse(input)
		if err != nil {
			return "", fmt.Errorf("chart parse error: %w", err)
		}
		return c.RenderHTML(), nil
	}

	return diagram.CreateDiagram(input)
}

// CreateDiagramWithSize generates an SVG diagram and returns it along with
// the computed canvas size in pixels. This is a package-level wrapper so callers
// needing the canvas size do not have to import pkg/diagram directly.
func CreateDiagramWithSize(code string) (string, int, int, error) {
	return diagram.CreateDiagramWithSize(code)
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
