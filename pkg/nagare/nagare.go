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

// RenderToSVGWithDebug is like RenderToSVG but prints debug information to stdout.
// func RenderToSVGWithDebug(code string) (string, error) {
// 	input := strings.TrimSpace(code)
// 	if strings.HasPrefix(input, "chart") {
// 		return RenderToSVG(code)
// 	}

// 	return diagram.RenderToSVGWithDebug(code)
// }

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
