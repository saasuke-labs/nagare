package chart

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DataPoint represents a single point in a series
type DataPoint struct {
	X float64
	Y float64
}

// Series represents a data series
type Series struct {
	Name  string
	Data  []DataPoint
	Color string
	Style string // "line", "dashed", "dotted"
}

// Chart represents a line chart
type Chart struct {
	Title      string
	Width      float64
	Height     float64
	XAxisType  string // "number" or "date"
	XAxisLabel string
	YAxisLabel string
	Legend     string // "top-left", "top-right", "bottom-left", "bottom-right", "none"
	Grid       bool
	Series     []Series
}

// Parse parses chart definition into a Chart struct
// Format:
// chart
// title: My Chart
// width: 900
// height: 600
// xaxis: date
// xlabel: Date
// ylabel: Value
// legend: top-right
// grid: true
//
// series: distance
// color: #3b82f6
// style: line
// data:
//
//	2025-01-01: 3.01
//	2025-01-02: 3.5
//	2025-01-03: 4.2
//
// series: pace
// color: #ef4444
// style: dashed
// data:
//
//	2025-01-01: 6.5
//	2025-01-02: 6.3
//	2025-01-03: 6.1
func Parse(input string) (*Chart, error) {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("empty input")
	}

	// First line should be "chart"
	if strings.TrimSpace(lines[0]) != "chart" {
		return nil, fmt.Errorf("first line must be 'chart'")
	}

	chart := &Chart{
		Width:     900,
		Height:    600,
		XAxisType: "number",
		Legend:    "top-right",
		Grid:      true,
		Series:    []Series{},
	}

	var currentSeries *Series
	var inDataBlock bool
	var dataLines []string

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			// Empty line - end data block if we're in one
			if inDataBlock && len(dataLines) > 0 {
				if currentSeries != nil {
					currentSeries.Data = parseDataBlock(dataLines, chart.XAxisType)
					dataLines = nil
				}
				inDataBlock = false
			}
			continue
		}

		// Check if line starts with spaces (multiline data)
		if len(line) > 0 && line[0] == ' ' || line[0] == '\t' {
			if !inDataBlock {
				continue // Skip indented lines outside data block
			}
			dataLines = append(dataLines, trimmed)
			continue
		}

		// Parse key: value
		if inDataBlock && len(dataLines) > 0 {
			if currentSeries != nil {
				currentSeries.Data = parseDataBlock(dataLines, chart.XAxisType)
				dataLines = nil
			}
			inDataBlock = false
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "title":
			chart.Title = value
		case "width":
			if w, err := strconv.ParseFloat(value, 64); err == nil {
				chart.Width = w
			}
		case "height":
			if h, err := strconv.ParseFloat(value, 64); err == nil {
				chart.Height = h
			}
		case "xaxis":
			chart.XAxisType = value
		case "xlabel":
			chart.XAxisLabel = value
		case "ylabel":
			chart.YAxisLabel = value
		case "legend":
			chart.Legend = value
		case "grid":
			chart.Grid = value == "true"
		case "series":
			// Save previous series if exists
			if currentSeries != nil && len(currentSeries.Data) > 0 {
				chart.Series = append(chart.Series, *currentSeries)
			}
			// Start new series
			currentSeries = &Series{
				Name:  value,
				Style: "line",
				Color: "",
			}
		case "color":
			if currentSeries != nil {
				currentSeries.Color = value
			}
		case "style":
			if currentSeries != nil {
				currentSeries.Style = value
			}
		case "data":
			// Start data block
			inDataBlock = true
			dataLines = []string{}
		}
	}

	// Handle final series
	if inDataBlock && currentSeries != nil && len(dataLines) > 0 {
		currentSeries.Data = parseDataBlock(dataLines, chart.XAxisType)
	}
	if currentSeries != nil && len(currentSeries.Data) > 0 {
		chart.Series = append(chart.Series, *currentSeries)
	}

	// Assign default colors
	defaultColors := []string{"#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899"}
	for i := range chart.Series {
		if chart.Series[i].Color == "" {
			chart.Series[i].Color = defaultColors[i%len(defaultColors)]
		}
	}

	// Sort series by name for consistency
	sort.Slice(chart.Series, func(i, j int) bool {
		return chart.Series[i].Name < chart.Series[j].Name
	})

	return chart, nil
}

// parseDataBlock parses multiline data in format:
//
//	x1: y1
//	x2: y2
//	x3: y3
func parseDataBlock(lines []string, xAxisType string) []DataPoint {
	points := []DataPoint{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		xStr := strings.TrimSpace(parts[0])
		yStr := strings.TrimSpace(parts[1])

		var x, y float64
		var err error

		// Parse X based on type
		if xAxisType == "date" {
			t, parseErr := time.Parse("2006-01-02", xStr)
			if parseErr == nil {
				x = float64(t.Unix())
			} else {
				continue
			}
		} else {
			x, err = strconv.ParseFloat(xStr, 64)
			if err != nil {
				continue
			}
		}

		// Parse Y
		y, err = strconv.ParseFloat(yStr, 64)
		if err != nil {
			continue
		}

		points = append(points, DataPoint{X: x, Y: y})
	}

	// Sort by X
	sort.Slice(points, func(i, j int) bool {
		return points[i].X < points[j].X
	})

	return points
}

// RenderSVG generates SVG output
func (c *Chart) RenderSVG() string {
	if len(c.Series) == 0 {
		return fmt.Sprintf(`<svg width="%.0f" height="%.0f"><text x="%.0f" y="%.0f" fill="#999">No data</text></svg>`,
			c.Width, c.Height, c.Width/2, c.Height/2)
	}

	padding := 60.0
	topPadding := 40.0
	if c.Title != "" {
		topPadding = 60.0
	}

	plotWidth := c.Width - 2*padding
	plotHeight := c.Height - padding - topPadding
	plotX := padding
	plotY := topPadding

	xMin, xMax, yMin, yMax := c.calculateRanges()

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%.0f" height="%.0f" xmlns="http://www.w3.org/2000/svg">`, c.Width, c.Height))
	svg.WriteString(`<rect width="100%" height="100%" fill="white"/>`)

	// Title
	if c.Title != "" {
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="25" text-anchor="middle" font-size="16" font-weight="bold" fill="#333">%s</text>`,
			c.Width/2, c.Title))
	}

	// Grid
	if c.Grid {
		svg.WriteString(c.generateGrid(plotX, plotY, plotWidth, plotHeight))
	}

	// Axes
	svg.WriteString(c.generateAxes(plotX, plotY, plotWidth, plotHeight, xMin, xMax, yMin, yMax))

	// Series
	svg.WriteString(c.generateSeries(plotX, plotY, plotWidth, plotHeight, xMin, xMax, yMin, yMax))

	// Legend
	if c.Legend != "none" && len(c.Series) > 1 {
		svg.WriteString(c.generateLegend(plotX, plotY, plotWidth, plotHeight))
	}

	svg.WriteString(`</svg>`)
	return svg.String()
}

func (c *Chart) calculateRanges() (xMin, xMax, yMin, yMax float64) {
	xMin, yMin = math.MaxFloat64, math.MaxFloat64
	xMax, yMax = -math.MaxFloat64, -math.MaxFloat64

	for _, series := range c.Series {
		for _, point := range series.Data {
			if point.X < xMin {
				xMin = point.X
			}
			if point.X > xMax {
				xMax = point.X
			}
			if point.Y < yMin {
				yMin = point.Y
			}
			if point.Y > yMax {
				yMax = point.Y
			}
		}
	}

	// Add 10% padding
	xRange := xMax - xMin
	yRange := yMax - yMin

	if xRange == 0 {
		xRange = 1
	}
	if yRange == 0 {
		yRange = 1
	}

	xMin -= xRange * 0.1
	xMax += xRange * 0.1
	yMin -= yRange * 0.1
	yMax += yRange * 0.1

	return
}

func (c *Chart) generateGrid(plotX, plotY, plotWidth, plotHeight float64) string {
	var lines strings.Builder

	// Vertical lines
	for i := 0; i <= 5; i++ {
		x := plotX + float64(i)*plotWidth/5.0
		lines.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#e5e7eb" stroke-width="1"/>`,
			x, plotY, x, plotY+plotHeight))
	}

	// Horizontal lines
	for i := 0; i <= 5; i++ {
		y := plotY + float64(i)*plotHeight/5.0
		lines.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#e5e7eb" stroke-width="1"/>`,
			plotX, y, plotX+plotWidth, y))
	}

	return lines.String()
}

func (c *Chart) generateAxes(plotX, plotY, plotWidth, plotHeight, xMin, xMax, yMin, yMax float64) string {
	var svg strings.Builder

	// X axis
	svg.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#333" stroke-width="2"/>`,
		plotX, plotY+plotHeight, plotX+plotWidth, plotY+plotHeight))

	// Y axis
	svg.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#333" stroke-width="2"/>`,
		plotX, plotY, plotX, plotY+plotHeight))

	// X axis labels
	for i := 0; i <= 5; i++ {
		x := plotX + float64(i)*plotWidth/5.0
		value := xMin + float64(i)*(xMax-xMin)/5.0

		label := ""
		if c.XAxisType == "date" {
			t := time.Unix(int64(value), 0)
			label = t.Format("01/02")
		} else {
			label = fmt.Sprintf("%.1f", value)
		}

		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="11" fill="#666">%s</text>`,
			x, plotY+plotHeight+20, label))
	}

	// Y axis labels
	for i := 0; i <= 5; i++ {
		y := plotY + plotHeight - float64(i)*plotHeight/5.0
		value := yMin + float64(i)*(yMax-yMin)/5.0

		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="end" font-size="11" fill="#666">%.1f</text>`,
			plotX-10, y+4, value))
	}

	// Axis labels
	if c.XAxisLabel != "" {
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12" fill="#333">%s</text>`,
			plotX+plotWidth/2, plotY+plotHeight+40, c.XAxisLabel))
	}

	if c.YAxisLabel != "" {
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12" fill="#333" transform="rotate(-90 %.2f %.2f)">%s</text>`,
			plotX-40, plotY+plotHeight/2, plotX-40, plotY+plotHeight/2, c.YAxisLabel))
	}

	return svg.String()
}

func (c *Chart) generateSeries(plotX, plotY, plotWidth, plotHeight, xMin, xMax, yMin, yMax float64) string {
	var svg strings.Builder

	for _, series := range c.Series {
		if len(series.Data) == 0 {
			continue
		}

		// Build path
		var pathData strings.Builder
		for i, point := range series.Data {
			x := plotX + (point.X-xMin)/(xMax-xMin)*plotWidth
			y := plotY + plotHeight - (point.Y-yMin)/(yMax-yMin)*plotHeight

			if i == 0 {
				pathData.WriteString(fmt.Sprintf("M %.2f %.2f", x, y))
			} else {
				pathData.WriteString(fmt.Sprintf(" L %.2f %.2f", x, y))
			}
		}

		// Stroke style
		strokeDasharray := ""
		if series.Style == "dashed" {
			strokeDasharray = ` stroke-dasharray="5,5"`
		} else if series.Style == "dotted" {
			strokeDasharray = ` stroke-dasharray="2,2"`
		}

		// Draw line
		svg.WriteString(fmt.Sprintf(`<path d="%s" stroke="%s" stroke-width="2" fill="none"%s/>`,
			pathData.String(), series.Color, strokeDasharray))

		// Draw points
		for _, point := range series.Data {
			x := plotX + (point.X-xMin)/(xMax-xMin)*plotWidth
			y := plotY + plotHeight - (point.Y-yMin)/(yMax-yMin)*plotHeight

			svg.WriteString(fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="3" fill="%s"/>`,
				x, y, series.Color))
		}
	}

	return svg.String()
}

func (c *Chart) generateLegend(plotX, plotY, plotWidth, plotHeight float64) string {
	var svg strings.Builder

	legendX, legendY := plotX+plotWidth-120, plotY+10

	if strings.Contains(c.Legend, "left") {
		legendX = plotX + 10
	}
	if strings.Contains(c.Legend, "bottom") {
		legendY = plotY + plotHeight - float64(len(c.Series))*25 - 10
	}

	legendHeight := float64(len(c.Series)) * 20
	svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="110" height="%.2f" fill="white" stroke="#ccc" stroke-width="1" rx="3"/>`,
		legendX, legendY, legendHeight+10))

	for i, series := range c.Series {
		y := legendY + 15 + float64(i)*20

		svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="15" height="3" fill="%s"/>`,
			legendX+5, y-2, series.Color))

		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" font-size="11" fill="#333">%s</text>`,
			legendX+25, y+2, series.Name))
	}

	return svg.String()
}

// RenderHTML wraps SVG in HTML
func (c *Chart) RenderHTML() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>%s</title>
</head>
<body style="margin: 0; padding: 20px; background: #f5f5f5;">
%s
</body>
</html>`, c.Title, c.RenderSVG())
}
