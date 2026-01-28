package chart

import (
	"strings"
	"testing"
)

func TestParseSimpleChart(t *testing.T) {
	input := `chart
title: Test Chart
xaxis: number

series: test
color: #ff0000
data:
  0: 10
  1: 20
  2: 15`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if c.Title != "Test Chart" {
		t.Errorf("Expected title 'Test Chart', got '%s'", c.Title)
	}

	if len(c.Series) != 1 {
		t.Fatalf("Expected 1 series, got %d", len(c.Series))
	}

	if len(c.Series[0].Data) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(c.Series[0].Data))
	}
}

func TestParseMultiSeries(t *testing.T) {
	input := `chart
title: Multi Series

series: a
color: #ff0000
data:
  0: 10
  1: 20

series: b
color: #00ff00
style: dashed
data:
  0: 5
  1: 15`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(c.Series) != 2 {
		t.Fatalf("Expected 2 series, got %d", len(c.Series))
	}
}

func TestParseMultiSeriesPipeDelimited(t *testing.T) {
	input := `chart
title: Compact Format

series: distance | pace
color: #3b82f6 | #ef4444
style: line | dashed
data:
  2025-01-01: 3.01 | 6.5
  2025-01-02: 3.5 | 6.3
  2025-01-03: 4.0 | 6.1`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(c.Series) != 2 {
		t.Fatalf("Expected 2 series, got %d", len(c.Series))
	}

	if c.Series[0].Name != "distance" && c.Series[0].Name != "pace" {
		t.Errorf("Expected 'distance' or 'pace' as series name, got '%s'", c.Series[0].Name)
	}

	if len(c.Series[0].Data) != 3 {
		t.Errorf("Expected 3 data points in first series, got %d", len(c.Series[0].Data))
	}

	if len(c.Series[1].Data) != 3 {
		t.Errorf("Expected 3 data points in second series, got %d", len(c.Series[1].Data))
	}

	// Check colors are applied correctly
	if c.Series[0].Color == "" {
		t.Error("Expected series 0 to have a color")
	}
	if c.Series[1].Color == "" {
		t.Error("Expected series 1 to have a color")
	}
}

func TestRenderSVG(t *testing.T) {
	input := `chart
title: Test

series: test
data:
  0: 10
  1: 20
  2: 15`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	svg := c.RenderSVG()

	if !strings.Contains(svg, "<svg") {
		t.Error("Expected SVG to contain <svg tag")
	}

	if !strings.Contains(svg, "Test") {
		t.Error("Expected SVG to contain title")
	}

	if !strings.Contains(svg, "<path") {
		t.Error("Expected SVG to contain path element")
	}
}
