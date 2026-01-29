package chart

import (
	"strings"
	"testing"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		valid    bool
	}{
		{"5:23", 5*60 + 23, true},
		{"33:01", 33*60 + 1, true},
		{"3:02:31", 3*3600 + 2*60 + 31, true},
		{"0:30", 30, true},
		{"1:00:00", 3600, true},
		{"invalid", 0, false},
		{"5", 0, false},
		{"5:23:12:45", 0, false},
	}

	for _, test := range tests {
		result, valid := parseDuration(test.input)
		if valid != test.valid {
			t.Errorf("parseDuration(%s) valid = %v, expected %v", test.input, valid, test.valid)
		}
		if valid && result != test.expected {
			t.Errorf("parseDuration(%s) = %.0f, expected %.0f", test.input, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{323, "5:23"},
		{1981, "33:01"},
		{10951, "3:02:31"},
		{30, "0:30"},
		{3600, "1:00:00"},
	}

	for _, test := range tests {
		result := formatDuration(test.input)
		if result != test.expected {
			t.Errorf("formatDuration(%.0f) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

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

func TestParseDurationChart(t *testing.T) {
	input := `chart
title: Running Times
xaxis: date

series: time
color: #3b82f6
type: duration
data:
  2025-01-01: 33:01
  2025-01-02: 5:23
  2025-01-03: 3:02:31`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(c.Series) != 1 {
		t.Fatalf("Expected 1 series, got %d", len(c.Series))
	}

	series := c.Series[0]
	if series.Type != "duration" {
		t.Errorf("Expected series type 'duration', got '%s'", series.Type)
	}

	if len(series.Data) != 3 {
		t.Fatalf("Expected 3 data points, got %d", len(series.Data))
	}

	// Check first point (33:01 = 1981 seconds)
	if series.Data[0].Y != 1981 {
		t.Errorf("Expected first point Y=1981, got %.0f", series.Data[0].Y)
	}
	if series.Data[0].YLabel != "33:01" {
		t.Errorf("Expected first point YLabel='33:01', got '%s'", series.Data[0].YLabel)
	}
	if !series.Data[0].YIsDuration {
		t.Error("Expected first point YIsDuration=true")
	}

	// Check second point (5:23 = 323 seconds)
	if series.Data[1].Y != 323 {
		t.Errorf("Expected second point Y=323, got %.0f", series.Data[1].Y)
	}

	// Check third point (3:02:31 = 10951 seconds)
	if series.Data[2].Y != 10951 {
		t.Errorf("Expected third point Y=10951, got %.0f", series.Data[2].Y)
	}
}

func TestParseMultiSeriesMixedTypes(t *testing.T) {
	input := `chart
title: Mixed Types

series: duration1 | duration2 | number
color: #3b82f6 | #ef4444 | #10b981
type: duration | duration | number
data:
  0: 5:30 | 10:15 | 42
  1: 6:00 | 9:45 | 38
  2: 5:45 | 10:00 | 40`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(c.Series) != 3 {
		t.Fatalf("Expected 3 series, got %d", len(c.Series))
	}

	// Check first series (duration)
	series0 := c.Series[0]
	if series0.Type != "duration" {
		t.Errorf("Expected series 0 type 'duration', got '%s'", series0.Type)
	}
	if len(series0.Data) != 3 {
		t.Errorf("Expected 3 data points in series 0, got %d", len(series0.Data))
	}
	if series0.Data[0].Y != 330 { // 5:30 = 330 seconds
		t.Errorf("Expected series 0 first point Y=330, got %.0f", series0.Data[0].Y)
	}

	// Check second series (duration)
	series1 := c.Series[1]
	if series1.Type != "duration" {
		t.Errorf("Expected series 1 type 'duration', got '%s'", series1.Type)
	}
	if len(series1.Data) != 3 {
		t.Errorf("Expected 3 data points in series 1, got %d", len(series1.Data))
	}
	if series1.Data[0].Y != 615 { // 10:15 = 615 seconds
		t.Errorf("Expected series 1 first point Y=615, got %.0f", series1.Data[0].Y)
	}

	// Check third series (number)
	series2 := c.Series[2]
	if series2.Type != "number" {
		t.Errorf("Expected series 2 type 'number', got '%s'", series2.Type)
	}
	if len(series2.Data) != 3 {
		t.Errorf("Expected 3 data points in series 2, got %d", len(series2.Data))
	}
	if series2.Data[0].Y != 42 {
		t.Errorf("Expected series 2 first point Y=42, got %.0f", series2.Data[0].Y)
	}
	if series2.Data[0].YIsDuration {
		t.Error("Expected series 2 first point YIsDuration=false")
	}
}

