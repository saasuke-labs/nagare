package chart

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestNiceStep(t *testing.T) {
	tests := []struct {
		raw  float64
		want float64
	}{
		{0.03, 0.05},
		{0.2, 0.2},
		{1, 1},
		{1.24, 2},
		{2.2, 2.5},
		{4.1, 5},
		{6.4, 10},
		{47, 50},
		{420, 500},
	}
	for _, tt := range tests {
		if got := niceStep(tt.raw); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("niceStep(%v) = %v, want %v", tt.raw, got, tt.want)
		}
	}
}

func TestNiceBounds(t *testing.T) {
	tests := []struct {
		name     string
		lo, hi   float64
		duration bool
		min, max float64
	}{
		{"positive data starts at zero", 0, 20.5, false, 0, 25},
		{"large values", 0, 235, false, 0, 250},
		{"negative lower bound", -3, 29, false, -10, 40},
		{"lower bound snap leaves the top uncovered", -3, 47, false, -20, 80},
		{"small fractions", 0, 0.51, false, 0, 1},
		{"flat series", 0, 0, false, 0, 1},
		{"durations use time steps", 0, 2100, true, 0, 3000},
		{"marathon durations", 0, 3*3600 + 45*60, true, 0, 5 * 3600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := niceBounds(tt.lo, tt.hi, tt.duration)
			if math.Abs(min-tt.min) > 1e-9 || math.Abs(max-tt.max) > 1e-9 {
				t.Errorf("niceBounds(%v, %v) = (%v, %v), want (%v, %v)", tt.lo, tt.hi, min, max, tt.min, tt.max)
			}
			if min > tt.lo || max < tt.hi {
				t.Errorf("bounds (%v, %v) do not cover the data (%v, %v)", min, max, tt.lo, tt.hi)
			}
		})
	}
}

func TestFormatTick(t *testing.T) {
	tests := []struct {
		value, step float64
		want        string
	}{
		{20, 5, "20"},
		{-10, 10, "-10"},
		{0, 0.2, "0"},
		{0.6000000000000001, 0.2, "0.6"},
		{0.25, 0.25, "0.25"},
		{7.5, 2.5, "7.5"},
		{10, 2.5, "10"},
		{-0.0000001, 5, "0"},
	}
	for _, tt := range tests {
		if got := formatTick(tt.value, tt.step); got != tt.want {
			t.Errorf("formatTick(%v, %v) = %q, want %q", tt.value, tt.step, got, tt.want)
		}
	}
}

// xLabels returns the X axis labels of a rendered chart with their positions.
func xLabels(t *testing.T, svg string) (labels []string, positions []float64) {
	t.Helper()
	re := regexp.MustCompile(`<text x="([0-9.]+)" y="[0-9.]+" text-anchor="middle" font-size="11" fill="#666">([^<]*)</text>`)
	for _, m := range re.FindAllStringSubmatch(svg, -1) {
		pos, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatalf("bad label position %q: %v", m[1], err)
		}
		positions = append(positions, pos)
		labels = append(labels, m[2])
	}
	return labels, positions
}

// yLabels returns the left Y axis labels of a rendered chart, bottom to top.
func yLabels(svg string) []string {
	re := regexp.MustCompile(`text-anchor="end" font-size="11" fill="#666">([^<]*)</text>`)
	var labels []string
	for _, m := range re.FindAllStringSubmatch(svg, -1) {
		labels = append(labels, m[1])
	}
	return labels
}

// pointPositions returns the x position of every data point marker.
func pointPositions(t *testing.T, svg string) []float64 {
	t.Helper()
	re := regexp.MustCompile(`<circle cx="([0-9.]+)"`)
	var positions []float64
	for _, m := range re.FindAllStringSubmatch(svg, -1) {
		pos, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatalf("bad point position %q: %v", m[1], err)
		}
		positions = append(positions, pos)
	}
	return positions
}

func render(t *testing.T, input string) string {
	t.Helper()
	c, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	return c.RenderSVG()
}

func TestDailySeriesLabelsEveryPointWhereItIs(t *testing.T) {
	var input strings.Builder
	input.WriteString("chart\nwidth: 760\nheight: 300\nxaxis: date\n\nseries: high\ndata:\n")
	for day := 15; day <= 24; day++ {
		fmt.Fprintf(&input, "  2026-11-%d: %d\n", day, day)
	}
	svg := render(t, input.String())

	labels, positions := xLabels(t, svg)
	want := []string{"11/15", "11/16", "11/17", "11/18", "11/19", "11/20", "11/21", "11/22", "11/23", "11/24"}
	if strings.Join(labels, " ") != strings.Join(want, " ") {
		t.Fatalf("X labels = %v, want %v", labels, want)
	}

	points := pointPositions(t, svg)
	if len(points) != len(positions) {
		t.Fatalf("expected one marker per label, got %d markers and %d labels", len(points), len(positions))
	}
	for i := range positions {
		if math.Abs(positions[i]-points[i]) > 0.02 {
			t.Errorf("label %q is at x=%.2f but its data point is at x=%.2f", labels[i], positions[i], points[i])
		}
	}
}

func TestMonthlySeriesUsesMonthNames(t *testing.T) {
	var input strings.Builder
	input.WriteString("chart\nwidth: 760\nheight: 340\nxaxis: date\n\nseries: rain\nstyle: bar\ndata:\n")
	for month := 1; month <= 12; month++ {
		fmt.Fprintf(&input, "  2025-%02d-15: %d\n", month, 30+month)
	}
	labels, _ := xLabels(t, render(t, input.String()))

	want := "Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec"
	if strings.Join(labels, " ") != want {
		t.Errorf("X labels = %v, want %s", labels, want)
	}
}

func TestLongDateSeriesUsesCalendarTicks(t *testing.T) {
	var input strings.Builder
	input.WriteString("chart\nwidth: 900\nheight: 400\nxaxis: date\n\nseries: visits\ndata:\n")
	year, month := 2023, 3
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&input, "  %d-%02d-01: %d\n", year, month, 1000+i*10)
		month++
		if month > 12 {
			month, year = 1, year+1
		}
	}
	labels, positions := xLabels(t, render(t, input.String()))

	// Quarterly ticks; on a multi-year axis January shows the year.
	want := "Apr Jul Oct 2024 Apr Jul Oct 2025 Apr Jul Oct 2026 Apr"
	if strings.Join(labels, " ") != want {
		t.Errorf("X labels = %v, want %s", labels, want)
	}
	for i := 1; i < len(positions); i++ {
		if positions[i]-positions[i-1] < minXTickSpacing*0.6 {
			t.Errorf("labels %q and %q overlap", labels[i-1], labels[i])
		}
	}
}

func TestNumberAxisTicks(t *testing.T) {
	t.Run("few points get one integer label each", func(t *testing.T) {
		var input strings.Builder
		input.WriteString("chart\nwidth: 760\nheight: 300\nxaxis: number\n\nseries: sales\ndata:\n")
		for month := 1; month <= 12; month++ {
			fmt.Fprintf(&input, "  %d: %d\n", month, month*3)
		}
		labels, _ := xLabels(t, render(t, input.String()))
		if strings.Join(labels, " ") != "1 2 3 4 5 6 7 8 9 10 11 12" {
			t.Errorf("X labels = %v", labels)
		}
	})

	t.Run("many points get round steps", func(t *testing.T) {
		var input strings.Builder
		input.WriteString("chart\nwidth: 900\nheight: 300\nxaxis: number\n\nseries: latency\ndata:\n")
		for i := 0; i < 200; i++ {
			fmt.Fprintf(&input, "  %.1f: %d\n", float64(i)*0.5, i%7)
		}
		labels, _ := xLabels(t, render(t, input.String()))
		if strings.Join(labels, " ") != "0 10 20 30 40 50 60 70 80 90" {
			t.Errorf("X labels = %v", labels)
		}
	})
}

func TestAutoScaleLabelsAreRound(t *testing.T) {
	svg := render(t, `chart
xaxis: date

series: high
data:
  2026-11-15: 18.1
  2026-11-16: 20
  2026-11-17: 16.5`)

	if got := strings.Join(yLabels(svg), " "); got != "0 5 10 15 20 25" {
		t.Errorf("Y labels = %q, want %q", got, "0 5 10 15 20 25")
	}
}

func TestManualScaleKeepsItsRange(t *testing.T) {
	svg := render(t, `chart
xaxis: number

scale
  id: temp
  min: -10
  max: 40

series: high
yaxis: temp
data:
  1: 12
  2: 31`)

	if got := strings.Join(yLabels(svg), " "); got != "-10 0 10 20 30 40" {
		t.Errorf("Y labels = %q, want %q", got, "-10 0 10 20 30 40")
	}
}

func TestImplicitScaleShowsDurations(t *testing.T) {
	svg := render(t, `chart
xaxis: date

series: 5K
type: duration
data:
  2025-01-01: 25:30
  2025-01-08: 24:45`)

	if got := strings.Join(yLabels(svg), " "); got != "0:00 10:00 20:00 30:00 40:00 50:00" {
		t.Errorf("Y labels = %q, want %q", got, "0:00 10:00 20:00 30:00 40:00 50:00")
	}
}

func TestRenderSVGHasViewBox(t *testing.T) {
	svg := render(t, `chart
width: 760
height: 300

series: a
data:
  1: 1
  2: 2`)

	if !strings.HasPrefix(svg, `<svg width="760" height="300" viewBox="0 0 760 300"`) {
		t.Errorf("expected width, height and a matching viewBox on the root element, got %q", svg[:80])
	}
}
