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
	X         float64
	Y         float64
	YLabel    string // Original label for Y value (e.g., duration format)
	YIsDuration bool // Whether Y is a duration value
}

// Series represents a data series
type Series struct {
	Name  string
	Data  []DataPoint
	Color string
	Style string // "line", "dashed", "dotted"
	Type  string // "number" or "duration"
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
// Supports two formats:
//
// Format 1: Single series per block (classic)
// chart
// series: distance
// color: #3b82f6
// data:
//
//	2025-01-01: 3.01
//	2025-01-02: 3.5
//
// series: pace
// color: #ef4444
// data:
//
//	2025-01-01: 6.5
//	2025-01-02: 6.3
//
// Format 2: Multiple series in one block (compact)
// chart
// series: distance | pace
// color: #3b82f6 | #ef4444
// style: line | dashed
// data:
//
//	2025-01-01: 3.01 | 6.5
//	2025-01-02: 3.5 | 6.3
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

	var seriesNames []string
	var seriesColors []string
	var seriesStyles []string
	var seriesTypes []string
	var inDataBlock bool
	var dataLines []string

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			// Empty line - end data block if we're in one
			if inDataBlock && len(dataLines) > 0 {
				processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, dataLines, chart.XAxisType)
				dataLines = nil
				seriesNames = nil
				seriesColors = nil
				seriesStyles = nil
				seriesTypes = nil
			}
			inDataBlock = false
			continue
		}

		// Check if line starts with spaces (multiline data)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if !inDataBlock {
				continue // Skip indented lines outside data block
			}
			dataLines = append(dataLines, trimmed)
			continue
		}

		// Parse key: value
		if inDataBlock && len(dataLines) > 0 {
			processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, dataLines, chart.XAxisType)
			dataLines = nil
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
			// Check if pipe-separated (multi-series format)
			if strings.Contains(value, "|") {
				seriesNames = parseDelimitedList(value, "|")
			} else {
				seriesNames = []string{value}
			}
		case "color":
			// Check if pipe-separated
			if strings.Contains(value, "|") {
				seriesColors = parseDelimitedList(value, "|")
			} else {
				seriesColors = []string{value}
			}
		case "style":
			// Check if pipe or comma-separated
			if strings.Contains(value, "|") {
				seriesStyles = parseDelimitedList(value, "|")
			} else if strings.Contains(value, ",") {
				seriesStyles = parseDelimitedList(value, ",")
			} else {
				seriesStyles = []string{value}
			}
		case "type":
			// Check if pipe-separated
			if strings.Contains(value, "|") {
				seriesTypes = parseDelimitedList(value, "|")
			} else {
				seriesTypes = []string{value}
			}
		case "data":
			// Start data block
			inDataBlock = true
			dataLines = []string{}
		}
	}

	// Handle final data block
	if inDataBlock && len(dataLines) > 0 {
		processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, dataLines, chart.XAxisType)
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

// parseDelimitedList splits a string by a delimiter and trims each element
func parseDelimitedList(value, delimiter string) []string {
	parts := strings.Split(value, delimiter)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// processMultiSeriesData handles both single and multi-series data blocks
func processMultiSeriesData(chart *Chart, seriesNames, seriesColors, seriesStyles, seriesTypes []string, dataLines []string, xAxisType string) {
	if len(seriesNames) == 0 {
		return
	}

	// Check if data is multi-series (pipe-separated values)
	isMultiSeries := false
	if len(dataLines) > 0 {
		firstLine := strings.TrimSpace(dataLines[0])
		parts := strings.SplitN(firstLine, ":", 2) // Use SplitN to split only on first colon
		if len(parts) == 2 {
			values := strings.Split(strings.TrimSpace(parts[1]), "|")
			isMultiSeries = len(values) > 1
		}
	}

	if isMultiSeries {
		// Parse multi-series data
		seriesDataMap := make(map[int][][2]string)

		for _, dataLine := range dataLines {
			dataLine = strings.TrimSpace(dataLine)
			parts := strings.SplitN(dataLine, ":", 2)
			if len(parts) != 2 {
				continue
			}

			xStr := strings.TrimSpace(parts[0])
			values := strings.Split(strings.TrimSpace(parts[1]), "|")

			for idx, valueStr := range values {
				valueStr = strings.TrimSpace(valueStr)
				seriesDataMap[idx] = append(seriesDataMap[idx], [2]string{xStr, valueStr})
			}
		}

		// Create series from the map
		for idx, dataPoints := range seriesDataMap {
			if idx >= len(seriesNames) {
				break
			}

			// Get the type for this series
			dataType := "number"
			if idx < len(seriesTypes) {
				dataType = seriesTypes[idx]
			}

			series := Series{
				Name: seriesNames[idx],
				Data: parseDataPoints(dataPoints, xAxisType, dataType),
				Type: dataType,
			}

			if idx < len(seriesColors) {
				series.Color = seriesColors[idx]
			}

			if idx < len(seriesStyles) {
				series.Style = seriesStyles[idx]
			} else {
				series.Style = "line"
			}

			if series.Color == "" {
				series.Color = "#3b82f6"
			}

			chart.Series = append(chart.Series, series)
		}
	} else {
		// Single series data
		dataType := "number"
		if len(seriesTypes) > 0 {
			dataType = seriesTypes[0]
		}

		dataPoints := parseDataBlock(dataLines, xAxisType, dataType)

		if len(seriesNames) > 0 {
			series := Series{
				Name: seriesNames[0],
				Data: dataPoints,
				Type: dataType,
			}

			if len(seriesColors) > 0 {
				series.Color = seriesColors[0]
			}

			if len(seriesStyles) > 0 {
				series.Style = seriesStyles[0]
			} else {
				series.Style = "line"
			}

			if series.Color == "" {
				series.Color = "#3b82f6"
			}

			chart.Series = append(chart.Series, series)
		}
	}
}

// parseDuration parses duration strings in format MM:SS or HH:MM:SS
// Returns seconds and true if valid duration, otherwise 0 and false
func parseDuration(s string) (float64, bool) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false
	}

	var hours, minutes, seconds float64
	var err error

	if len(parts) == 2 {
		// MM:SS format
		minutes, err = strconv.ParseFloat(parts[0], 64)
		if err != nil || minutes < 0 {
			return 0, false
		}
		seconds, err = strconv.ParseFloat(parts[1], 64)
		if err != nil || seconds < 0 {
			return 0, false
		}
	} else {
		// HH:MM:SS format
		hours, err = strconv.ParseFloat(parts[0], 64)
		if err != nil || hours < 0 {
			return 0, false
		}
		minutes, err = strconv.ParseFloat(parts[1], 64)
		if err != nil || minutes < 0 {
			return 0, false
		}
		seconds, err = strconv.ParseFloat(parts[2], 64)
		if err != nil || seconds < 0 {
			return 0, false
		}
	}

	totalSeconds := hours*3600 + minutes*60 + seconds
	return totalSeconds, true
}

// formatDuration formats seconds back to duration string
func formatDuration(seconds float64) string {
	// Handle negative values by taking absolute value
	if seconds < 0 {
		seconds = 0
	}

	hours := int(seconds / 3600)
	minutes := int((seconds - float64(hours*3600)) / 60)
	secs := int(seconds - float64(hours*3600) - float64(minutes*60))

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// parseDataBlock parses multiline data in format:
//
//	x1: y1
//	x2: y2
func parseDataBlock(lines []string, xAxisType string, dataType string) []DataPoint {
	dataPoints := make([][2]string, 0, len(lines))

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

		dataPoints = append(dataPoints, [2]string{xStr, yStr})
	}

	return parseDataPoints(dataPoints, xAxisType, dataType)
}

// parseDataPoints converts string pairs to DataPoints
func parseDataPoints(dataPoints [][2]string, xAxisType string, dataType string) []DataPoint {
	points := []DataPoint{}

	for _, pair := range dataPoints {
		xStr := pair[0]
		yStr := pair[1]

		var x, y float64
		var err error
		var yLabel string
		var yIsDuration bool

		// Try to parse as date first, then fall back to numeric
		t, dateErr := time.Parse("2006-01-02", xStr)
		if dateErr == nil {
			x = float64(t.Unix())
		} else {
			// Try numeric parsing
			x, err = strconv.ParseFloat(xStr, 64)
			if err != nil {
				continue
			}
		}

		// Parse Y based on type
		if dataType == "duration" {
			// Try to parse as duration
			if seconds, ok := parseDuration(yStr); ok {
				y = seconds
				yLabel = yStr
				yIsDuration = true
			} else {
				// If duration parsing fails, skip this point
				continue
			}
		} else {
			// Parse as number
			y, err = strconv.ParseFloat(yStr, 64)
			if err != nil {
				continue
			}
			yLabel = ""
			yIsDuration = false
		}

		points = append(points, DataPoint{X: x, Y: y, YLabel: yLabel, YIsDuration: yIsDuration})
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
	// Check if any series has duration data
	hasDuration := false
	for _, series := range c.Series {
		if series.Type == "duration" {
			hasDuration = true
			break
		}
	}

	for i := 0; i <= 5; i++ {
		y := plotY + plotHeight - float64(i)*plotHeight/5.0
		value := yMin + float64(i)*(yMax-yMin)/5.0

		label := ""
		if hasDuration {
			label = formatDuration(value)
		} else {
			label = fmt.Sprintf("%.1f", value)
		}

		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="end" font-size="11" fill="#666">%s</text>`,
			plotX-10, y+4, label))
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
