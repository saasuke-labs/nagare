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
		{"", 0, false},
		{"  ", 0, false},
		{"-5:30", 0, false},
		{"5:-30", 0, false},
		{"-5:-30", 0, false},
		{"5: 30", 0, false},
		{"abc:def", 0, false},
		{"5.5:30", 5.5*60 + 30, true}, // Fractional minutes should work
	}

	for _, test := range tests {
		result, valid := parseDuration(test.input)
		if valid != test.valid {
			t.Errorf("parseDuration(%q) valid = %v, expected %v", test.input, valid, test.valid)
		}
		if valid && result != test.expected {
			t.Errorf("parseDuration(%q) = %.0f, expected %.0f", test.input, result, test.expected)
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
		{0, "0:00"},
		{-100, "0:00"}, // Negative values should be clamped to 0
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

func TestParseScaleDefinition(t *testing.T) {
	input := `chart
title: Multi-Scale Chart

scale
  id: time
  label: Duration (minutes)
  type: duration

scale
  id: distance
  label: Distance (km)

series: duration
yaxis: time
type: duration
data:
  1: 5:30
  2: 6:15

series: distance
yaxis: distance
data:
  1: 3.5
  2: 4.2`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check scales
	if len(c.Scales) != 2 {
		t.Fatalf("Expected 2 scales, got %d", len(c.Scales))
	}

	// Check first scale
	scale0 := c.Scales[0]
	if scale0.ID != "time" {
		t.Errorf("Expected scale 0 ID 'time', got '%s'", scale0.ID)
	}
	if scale0.Label != "Duration (minutes)" {
		t.Errorf("Expected scale 0 label 'Duration (minutes)', got '%s'", scale0.Label)
	}
	if scale0.Type != "duration" {
		t.Errorf("Expected scale 0 type 'duration', got '%s'", scale0.Type)
	}
	if !scale0.Auto {
		t.Error("Expected scale 0 Auto=true")
	}

	// Check second scale
	scale1 := c.Scales[1]
	if scale1.ID != "distance" {
		t.Errorf("Expected scale 1 ID 'distance', got '%s'", scale1.ID)
	}
	if scale1.Label != "Distance (km)" {
		t.Errorf("Expected scale 1 label 'Distance (km)', got '%s'", scale1.Label)
	}

	// Check series scale assignments
	if len(c.Series) != 2 {
		t.Fatalf("Expected 2 series, got %d", len(c.Series))
	}

	// Find series by name
	var durationSeries, distanceSeries *Series
	for i := range c.Series {
		if c.Series[i].Name == "duration" {
			durationSeries = &c.Series[i]
		} else if c.Series[i].Name == "distance" {
			distanceSeries = &c.Series[i]
		}
	}

	if durationSeries == nil {
		t.Fatal("Duration series not found")
	}
	if durationSeries.YAxis != "time" {
		t.Errorf("Expected duration series YAxis='time', got '%s'", durationSeries.YAxis)
	}

	if distanceSeries == nil {
		t.Fatal("Distance series not found")
	}
	if distanceSeries.YAxis != "distance" {
		t.Errorf("Expected distance series YAxis='distance', got '%s'", distanceSeries.YAxis)
	}
}

func TestMultiScaleRangeCalculation(t *testing.T) {
	input := `chart
title: Training Data

scale
  id: duration
  label: Time (min)
  type: duration

scale
  id: distance
  label: Distance (km)

series: run-time
yaxis: duration
type: duration
data:
  1: 30:00
  2: 35:00
  3: 32:00

series: run-distance
yaxis: distance
data:
  1: 5.0
  2: 6.2
  3: 5.5`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Calculate scale ranges
	c.calculateScaleRanges()

	if len(c.Scales) != 2 {
		t.Fatalf("Expected 2 scales, got %d", len(c.Scales))
	}

	// Check duration scale (30:00 = 1800s, 35:00 = 2100s)
	durationScale := c.Scales[0]
	if durationScale.Min >= 1800 {
		t.Errorf("Expected duration scale Min < 1800, got %.2f", durationScale.Min)
	}
	if durationScale.Max <= 2100 {
		t.Errorf("Expected duration scale Max > 2100, got %.2f", durationScale.Max)
	}

	// Check distance scale (5.0 to 6.2)
	distanceScale := c.Scales[1]
	if distanceScale.Min >= 5.0 {
		t.Errorf("Expected distance scale Min < 5.0, got %.2f", distanceScale.Min)
	}
	if distanceScale.Max <= 6.2 {
		t.Errorf("Expected distance scale Max > 6.2, got %.2f", distanceScale.Max)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that old charts without scale definitions still work
	input := `chart
title: Simple Chart
ylabel: Values

series: test
data:
  1: 10
  2: 20
  3: 15`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should create a default scale
	if len(c.Scales) != 1 {
		t.Fatalf("Expected 1 default scale, got %d", len(c.Scales))
	}

	if c.Scales[0].ID != "default" {
		t.Errorf("Expected default scale ID 'default', got '%s'", c.Scales[0].ID)
	}

	// Should use YAxisLabel for scale label
	if c.Scales[0].Label != "Values" {
		t.Errorf("Expected scale label 'Values', got '%s'", c.Scales[0].Label)
	}

	// Series should be assigned to default scale
	if c.Series[0].YAxis != "default" {
		t.Errorf("Expected series YAxis='default', got '%s'", c.Series[0].YAxis)
	}
}

func TestScaleWithManualRange(t *testing.T) {
	input := `chart
title: Chart with Manual Scale

scale
  id: custom
  label: Custom Scale
  min: 0
  max: 100

series: test
yaxis: custom
data:
  1: 30
  2: 50
  3: 70`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(c.Scales) != 1 {
		t.Fatalf("Expected 1 scale, got %d", len(c.Scales))
	}

	scale := c.Scales[0]
	if scale.Min != 0 {
		t.Errorf("Expected scale Min=0, got %.2f", scale.Min)
	}
	if scale.Max != 100 {
		t.Errorf("Expected scale Max=100, got %.2f", scale.Max)
	}
	if scale.Auto {
		t.Error("Expected scale Auto=false when min/max are set")
	}
}

func TestParseStackedBarAndMarkerSeries(t *testing.T) {
	input := `chart
title: Training Metrics
xaxis: date

scale
  id: time
  type: duration

scale
  id: distance

series: jog | walk | total | blocks
style: bar | bar | line | marker
stack: session | session | none | none
yaxis: time | time | distance | distance
type: duration | duration | number | number
data:
  2025-12-18: 3:51 | 25:04 | 2.95 | 1
  2025-12-19: 7:29 | 20:02 | 2.88 | 2`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(c.Series) != 4 {
		t.Fatalf("Expected 4 series, got %d", len(c.Series))
	}

	if c.Series[0].Style != "bar" || c.Series[1].Style != "bar" {
		t.Fatalf("Expected first two series to be bar style, got %q and %q", c.Series[0].Style, c.Series[1].Style)
	}

	if c.Series[0].Stack != "session" || c.Series[1].Stack != "session" {
		t.Fatalf("Expected stacked bar series to have stack=session, got %q and %q", c.Series[0].Stack, c.Series[1].Stack)
	}

	if c.Series[3].Style != "marker" {
		t.Fatalf("Expected marker series style, got %q", c.Series[3].Style)
	}

	if c.Series[0].Data[0].Y != 231 { // 3:51
		t.Fatalf("Expected duration to be converted to seconds for bars, got %.0f", c.Series[0].Data[0].Y)
	}
}

func TestRenderSVGStackedBarsAndMarkerAnnotations(t *testing.T) {
	input := `chart
title: Mixed
xaxis: date

scale
  id: time
  type: duration

scale
  id: distance

series: jog | walk | total | blocks
color: #aa7777 | #ddd0d0 | #7777aa | #666666
style: bar | bar | line | marker
stack: session | session | none | none
yaxis: time | time | distance | distance
type: duration | duration | number | number
data:
  2025-12-18: 3:51 | 25:04 | 2.95 | 1
  2025-12-19: 7:29 | 20:02 | 2.88 | 2`

	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	svg := c.RenderSVG()

	if strings.Count(svg, "<rect") < 3 {
		t.Fatalf("Expected bar rendering rects, got SVG: %s", svg)
	}

	if !strings.Contains(svg, `<text x=`) || !strings.Contains(svg, `>2</text>`) {
		t.Fatalf("Expected marker labels to render as text annotations, got SVG: %s", svg)
	}

	if !strings.Contains(svg, `stroke="#7777aa"`) {
		t.Fatalf("Expected line series stroke in SVG")
	}

	if !strings.Contains(svg, "jog") || !strings.Contains(svg, "walk") || !strings.Contains(svg, "total") || !strings.Contains(svg, "blocks") {
		t.Fatalf("Expected legend entries for mixed series")
	}
}
