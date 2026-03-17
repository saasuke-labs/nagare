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
	X           float64
	Y           float64
	YLabel      string // Original label for Y value (e.g., duration format)
	YIsDuration bool   // Whether Y is a duration value
}

// Scale represents a Y-axis scale with its own range and label
type Scale struct {
	ID    string  // Unique identifier (e.g., "default", "distance", "duration")
	Label string  // Label for this scale
	Min   float64 // Minimum value (used when Auto is false)
	Max   float64 // Maximum value (used when Auto is false)
	Auto  bool    // Whether to auto-calculate min/max
	Type  string  // "number" or "duration"
}

// Series represents a data series
type Series struct {
	Name  string
	Data  []DataPoint
	Color string
	Style string // "line", "dashed", "dotted", "bar", "marker"
	Type  string // "number" or "duration"
	YAxis string // Scale ID this series belongs to (default: "default")
	Stack string // Stack group for bar series ("none" disables stacking)
}

// Chart represents a line chart
type Chart struct {
	Title      string
	Width      float64
	Height     float64
	XAxisType  string // "number" or "date"
	XAxisLabel string
	YAxisLabel string // Deprecated: use Scales instead
	Legend     string // "top-left", "top-right", "bottom-left", "bottom-right", "none"
	Grid       bool
	Series     []Series
	Scales     []Scale // Y-axis scales
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
		Scales:    []Scale{},
	}

	var seriesNames []string
	var seriesColors []string
	var seriesStyles []string
	var seriesTypes []string
	var seriesYAxes []string
	var seriesStacks []string
	var inDataBlock bool
	var dataLines []string
	var currentScale *Scale
	var inScaleBlock bool
	var scaleMinSet, scaleMaxSet bool

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			// Empty line - end data/scale block if we're in one
			if inDataBlock && len(dataLines) > 0 {
				processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, seriesYAxes, seriesStacks, dataLines, chart.XAxisType)
				dataLines = nil
				seriesNames = nil
				seriesColors = nil
				seriesStyles = nil
				seriesTypes = nil
				seriesYAxes = nil
				seriesStacks = nil
			}
			if inScaleBlock && currentScale != nil {
				if scaleMinSet && scaleMaxSet {
					currentScale.Auto = false
				}
				chart.Scales = append(chart.Scales, *currentScale)
				currentScale = nil
				scaleMinSet = false
				scaleMaxSet = false
			}
			inDataBlock = false
			inScaleBlock = false
			continue
		}

		// Check if line starts with spaces (multiline data)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if !inDataBlock && !inScaleBlock {
				continue // Skip indented lines outside data/scale block
			}
			if inDataBlock {
				dataLines = append(dataLines, trimmed)
				continue
			}
			// For scale blocks, indented lines are properties - fall through to parsing
		}

		// Check for scale block start (non-indented "scale" keyword)
		if trimmed == "scale" && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			// End any previous blocks
			if inDataBlock && len(dataLines) > 0 {
				processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, seriesYAxes, seriesStacks, dataLines, chart.XAxisType)
				dataLines = nil
				inDataBlock = false
			}
			if inScaleBlock && currentScale != nil {
				// Finalize previous scale - set Auto based on whether min/max were set
				if scaleMinSet && scaleMaxSet {
					currentScale.Auto = false
				}
				chart.Scales = append(chart.Scales, *currentScale)
			}
			// Start new scale block
			inScaleBlock = true
			scaleMinSet = false
			scaleMaxSet = false
			currentScale = &Scale{
				Auto: true,
				Type: "number",
			}
			continue
		}

		// Parse key: value
		if inDataBlock && len(dataLines) > 0 {
			processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, seriesYAxes, seriesStacks, dataLines, chart.XAxisType)
			dataLines = nil
			inDataBlock = false
		}
		if inScaleBlock && currentScale != nil && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			// Non-indented line while in scale block - finalize and save scale
			if scaleMinSet && scaleMaxSet {
				currentScale.Auto = false
			}
			chart.Scales = append(chart.Scales, *currentScale)
			currentScale = nil
			inScaleBlock = false
			scaleMinSet = false
			scaleMaxSet = false
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
		case "yaxis":
			// Y-axis/scale assignment for series (only yaxis, not scale)
			if strings.Contains(value, "|") {
				seriesYAxes = parseDelimitedList(value, "|")
			} else {
				seriesYAxes = []string{value}
			}
		case "stack":
			if strings.Contains(value, "|") {
				seriesStacks = parseDelimitedList(value, "|")
			} else {
				seriesStacks = []string{value}
			}
		case "data":
			// Start data block
			inDataBlock = true
			dataLines = []string{}
		}

		// Handle scale block properties (when in scale block)
		if inScaleBlock && currentScale != nil {
			switch key {
			case "id":
				currentScale.ID = value
			case "label":
				currentScale.Label = value
			case "min":
				if min, err := strconv.ParseFloat(value, 64); err == nil {
					currentScale.Min = min
					scaleMinSet = true
				}
			case "max":
				if max, err := strconv.ParseFloat(value, 64); err == nil {
					currentScale.Max = max
					scaleMaxSet = true
				}
			case "type":
				currentScale.Type = value
			}
		}
	}

	// Handle final data block
	if inDataBlock && len(dataLines) > 0 {
		processMultiSeriesData(chart, seriesNames, seriesColors, seriesStyles, seriesTypes, seriesYAxes, seriesStacks, dataLines, chart.XAxisType)
	}

	// Handle final scale block
	if inScaleBlock && currentScale != nil {
		if scaleMinSet && scaleMaxSet {
			currentScale.Auto = false
		}
		chart.Scales = append(chart.Scales, *currentScale)
	}

	// Validate scales
	for i := range chart.Scales {
		scale := &chart.Scales[i]
		// If auto is disabled but range is invalid, reset to auto
		if !scale.Auto && scale.Min >= scale.Max {
			scale.Auto = true
			scale.Min = 0
			scale.Max = 0
		}
	}

	// Ensure there's at least a default scale if none defined
	if len(chart.Scales) == 0 {
		chart.Scales = append(chart.Scales, Scale{
			ID:   "default",
			Auto: true,
			Type: "number",
		})
		// Use old YAxisLabel if present
		if chart.YAxisLabel != "" {
			chart.Scales[0].Label = chart.YAxisLabel
		}
	}

	// Assign default yaxis to series that don't have one
	for i := range chart.Series {
		if chart.Series[i].YAxis == "" {
			chart.Series[i].YAxis = "default"
		}
	}

	// Assign default colors
	defaultColors := []string{"#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899"}
	for i := range chart.Series {
		if chart.Series[i].Color == "" {
			chart.Series[i].Color = defaultColors[i%len(defaultColors)]
		}
	}

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
func processMultiSeriesData(chart *Chart, seriesNames, seriesColors, seriesStyles, seriesTypes, seriesYAxes, seriesStacks []string, dataLines []string, xAxisType string) {
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
		// Parse multi-series data while preserving series declaration order
		seriesDataByIndex := make([][][2]string, len(seriesNames))

		for _, dataLine := range dataLines {
			dataLine = strings.TrimSpace(dataLine)
			parts := strings.SplitN(dataLine, ":", 2)
			if len(parts) != 2 {
				continue
			}

			xStr := strings.TrimSpace(parts[0])
			values := strings.Split(strings.TrimSpace(parts[1]), "|")

			for idx, valueStr := range values {
				if idx >= len(seriesDataByIndex) {
					break
				}
				valueStr = strings.TrimSpace(valueStr)
				seriesDataByIndex[idx] = append(seriesDataByIndex[idx], [2]string{xStr, valueStr})
			}
		}

		// Create series from collected data in declaration order
		for idx, dataPoints := range seriesDataByIndex {

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

			if idx < len(seriesYAxes) {
				series.YAxis = seriesYAxes[idx]
			}

			if idx < len(seriesStacks) {
				series.Stack = seriesStacks[idx]
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

			if len(seriesYAxes) > 0 {
				series.YAxis = seriesYAxes[0]
			}

			if len(seriesStacks) > 0 {
				series.Stack = seriesStacks[0]
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
	baseTopPadding := 40.0
	if c.Title != "" {
		baseTopPadding = 60.0
	}
	baseBottomPadding := 60.0
	leftPadding := padding
	rightPadding := padding

	// Check if we have multiple scales (need right axis)
	hasRightAxis := len(c.Scales) > 1
	if hasRightAxis {
		rightPadding = 60.0
	}

	legendLayout := legendLayout{}
	legendPos := c.legendPosition()
	showLegend := c.Legend != "none" && len(c.Series) > 1
	legendGap := 10.0
	if showLegend && legendPos != "none" {
		legendLayout.Orientation = legendPos
		legendLayout.Width, legendLayout.Height = c.legendSize(legendPos)
		switch legendPos {
		case "top":
			baseTopPadding += legendLayout.Height + legendGap
		case "bottom":
			baseBottomPadding += legendLayout.Height + legendGap
		case "left":
			leftPadding += legendLayout.Width + legendGap
		case "right":
			rightPadding += legendLayout.Width + legendGap
		}
	}

	plotWidth := c.Width - leftPadding - rightPadding
	plotHeight := c.Height - baseBottomPadding - baseTopPadding
	plotX := leftPadding
	plotY := baseTopPadding

	if showLegend && legendPos != "none" {
		switch legendPos {
		case "top":
			legendLayout.X = plotX + (plotWidth-legendLayout.Width)/2
			legendLayout.Y = plotY - legendLayout.Height - legendGap
		case "bottom":
			legendLayout.X = plotX + (plotWidth-legendLayout.Width)/2
			legendLayout.Y = plotY + plotHeight + legendGap
		case "left":
			legendLayout.X = plotX - legendLayout.Width - legendGap
			legendLayout.Y = plotY + (plotHeight-legendLayout.Height)/2
		case "right":
			legendLayout.X = plotX + plotWidth + legendGap
			legendLayout.Y = plotY + (plotHeight-legendLayout.Height)/2
		}
	}

	// Calculate X range (same for all series)
	xMin, xMax, _, _ := c.calculateRanges()

	// Calculate scale-specific Y ranges
	c.calculateScaleRanges()

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
	svg.WriteString(c.generateAxes(plotX, plotY, plotWidth, plotHeight, xMin, xMax))

	// Series
	svg.WriteString(c.generateSeries(plotX, plotY, plotWidth, plotHeight, xMin, xMax))

	// Legend
	if showLegend && legendPos != "none" {
		svg.WriteString(c.generateLegend(legendLayout))
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

	// Add padding (smaller on X so lines are closer to edges)
	xRange := xMax - xMin
	yRange := yMax - yMin

	if xRange == 0 {
		xRange = 1
	}
	if yRange == 0 {
		yRange = 1
	}

	xPaddingFactor := 0.03
	yPaddingFactor := 0.1
	xMin -= xRange * xPaddingFactor
	xMax += xRange * xPaddingFactor
	yMin -= yRange * yPaddingFactor
	yMax += yRange * yPaddingFactor

	return
}

type legendLayout struct {
	X, Y        float64
	Width       float64
	Height      float64
	Orientation string
}

func (c *Chart) legendPosition() string {
	legend := strings.ToLower(strings.TrimSpace(c.Legend))
	if legend == "none" || legend == "" {
		return "none"
	}
	if strings.Contains(legend, "top") {
		return "top"
	}
	if strings.Contains(legend, "bottom") {
		return "bottom"
	}
	if strings.Contains(legend, "left") {
		return "left"
	}
	if strings.Contains(legend, "right") {
		return "right"
	}
	return "top"
}

func (c *Chart) legendSize(position string) (width, height float64) {
	const (
		fontSize       = 11.0
		lineWidth      = 15.0
		textGap        = 6.0
		itemGap        = 14.0
		padding        = 8.0
		rowHeight      = 20.0
		charWidthRatio = 0.6
	)

	maxTextWidth := 0.0
	itemCount := float64(len(c.Series))
	for _, series := range c.Series {
		textWidth := float64(len([]rune(series.Name))) * fontSize * charWidthRatio
		if textWidth > maxTextWidth {
			maxTextWidth = textWidth
		}
	}

	itemWidth := lineWidth + textGap + maxTextWidth

	switch position {
	case "left", "right":
		width = padding*2 + itemWidth
		height = padding*2 + rowHeight*itemCount
	default: // top or bottom
		totalItemsWidth := 0.0
		for i, series := range c.Series {
			textWidth := float64(len([]rune(series.Name))) * fontSize * charWidthRatio
			itemW := lineWidth + textGap + textWidth
			totalItemsWidth += itemW
			if i < len(c.Series)-1 {
				totalItemsWidth += itemGap
			}
		}
		width = padding*2 + totalItemsWidth
		height = padding*2 + rowHeight
	}

	return width, height
}

// calculateScaleRanges calculates min/max for each scale based on its series
func (c *Chart) calculateScaleRanges() {
	for i := range c.Scales {
		scale := &c.Scales[i]

		// Skip if min/max are manually set
		if !scale.Auto {
			continue
		}

		yMin := math.MaxFloat64
		yMax := -math.MaxFloat64
		foundData := false
		stackedSumsByX := map[string]float64{}
		stackedNegativeSumsByX := map[string]float64{}

		// Find all series belonging to this scale
		for _, series := range c.Series {
			if series.YAxis != scale.ID {
				continue
			}

			isBar := series.Style == "bar"
			isStackedBar := isBar && series.Stack != "" && strings.ToLower(series.Stack) != "none"
			for _, point := range series.Data {
				foundData = true
				if isStackedBar {
					stackKey := fmt.Sprintf("%s:%f", series.Stack, point.X)
					if point.Y >= 0 {
						stackedSumsByX[stackKey] += point.Y
						if stackedSumsByX[stackKey] > yMax {
							yMax = stackedSumsByX[stackKey]
						}
					} else {
						stackedNegativeSumsByX[stackKey] += point.Y
						if stackedNegativeSumsByX[stackKey] < yMin {
							yMin = stackedNegativeSumsByX[stackKey]
						}
					}
					continue
				}

				if point.Y < yMin {
					yMin = point.Y
				}
				if point.Y > yMax {
					yMax = point.Y
				}
			}
		}

		// Handle case where no data was found for this scale
		if !foundData {
			scale.Min = 0
			scale.Max = 1
			continue
		}

		if yMin > 0 {
			yMin = 0
		}

		// Add 10% padding
		yRange := yMax - yMin
		if yRange == 0 {
			yRange = 1
		}

		scale.Min = yMin - yRange*0.1
		scale.Max = yMax + yRange*0.1
	}
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

func (c *Chart) generateAxes(plotX, plotY, plotWidth, plotHeight, xMin, xMax float64) string {
	var svg strings.Builder

	// X axis
	svg.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#333" stroke-width="2"/>`,
		plotX, plotY+plotHeight, plotX+plotWidth, plotY+plotHeight))

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

	// X axis label
	if c.XAxisLabel != "" {
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12" fill="#333">%s</text>`,
			plotX+plotWidth/2, plotY+plotHeight+40, c.XAxisLabel))
	}

	// Render Y axes based on scales
	if len(c.Scales) == 0 {
		return svg.String()
	}

	// Left Y axis (first scale)
	leftScale := c.Scales[0]
	svg.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#333" stroke-width="2"/>`,
		plotX, plotY, plotX, plotY+plotHeight))

	for i := 0; i <= 5; i++ {
		y := plotY + plotHeight - float64(i)*plotHeight/5.0
		value := leftScale.Min + float64(i)*(leftScale.Max-leftScale.Min)/5.0

		label := ""
		if leftScale.Type == "duration" {
			label = formatDuration(value)
		} else {
			label = fmt.Sprintf("%.1f", value)
		}

		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="end" font-size="11" fill="#666">%s</text>`,
			plotX-10, y+4, label))
	}

	// Left Y axis label
	if leftScale.Label != "" {
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12" fill="#333" transform="rotate(-90 %.2f %.2f)">%s</text>`,
			plotX-40, plotY+plotHeight/2, plotX-40, plotY+plotHeight/2, leftScale.Label))
	} else if c.YAxisLabel != "" {
		// Backward compatibility
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12" fill="#333" transform="rotate(-90 %.2f %.2f)">%s</text>`,
			plotX-40, plotY+plotHeight/2, plotX-40, plotY+plotHeight/2, c.YAxisLabel))
	}

	// Right Y axis (second scale if exists)
	if len(c.Scales) > 1 {
		rightScale := c.Scales[1]
		rightX := plotX + plotWidth

		svg.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#333" stroke-width="2"/>`,
			rightX, plotY, rightX, plotY+plotHeight))

		for i := 0; i <= 5; i++ {
			y := plotY + plotHeight - float64(i)*plotHeight/5.0
			value := rightScale.Min + float64(i)*(rightScale.Max-rightScale.Min)/5.0

			label := ""
			if rightScale.Type == "duration" {
				label = formatDuration(value)
			} else {
				label = fmt.Sprintf("%.1f", value)
			}

			svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="start" font-size="11" fill="#666">%s</text>`,
				rightX+10, y+4, label))
		}

		// Right Y axis label
		if rightScale.Label != "" {
			svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12" fill="#333" transform="rotate(90 %.2f %.2f)">%s</text>`,
				rightX+40, plotY+plotHeight/2, rightX+40, plotY+plotHeight/2, rightScale.Label))
		}
	}

	return svg.String()
}

func (c *Chart) generateSeries(plotX, plotY, plotWidth, plotHeight, xMin, xMax float64) string {
	var svg strings.Builder

	// Create a map of scale ID to scale for quick lookup
	scaleMap := make(map[string]*Scale)
	for i := range c.Scales {
		scaleMap[c.Scales[i].ID] = &c.Scales[i]
	}

	type barGroup struct {
		Key    string
		Series []Series
	}

	barGroupsByAxis := map[string][]barGroup{}
	for _, series := range c.Series {
		if series.Style != "bar" {
			continue
		}
		axis := series.YAxis
		if axis == "" {
			axis = "default"
		}
		groupKey := "series:" + series.Name
		if series.Stack != "" && strings.ToLower(series.Stack) != "none" {
			groupKey = "stack:" + series.Stack
		}

		groups := barGroupsByAxis[axis]
		found := false
		for gi := range groups {
			if groups[gi].Key == groupKey {
				groups[gi].Series = append(groups[gi].Series, series)
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, barGroup{Key: groupKey, Series: []Series{series}})
		}
		barGroupsByAxis[axis] = groups
	}

	for axisID, groups := range barGroupsByAxis {
		scale, ok := scaleMap[axisID]
		if !ok {
			if len(c.Scales) == 0 {
				continue
			}
			scale = &c.Scales[0]
		}

		yMin := scale.Min
		yMax := scale.Max
		if yMax == yMin {
			yMax = yMin + 1
		}

		seriesWithData := 0
		for _, group := range groups {
			for _, series := range group.Series {
				if len(series.Data) > 0 {
					seriesWithData++
					break
				}
			}
		}
		if seriesWithData == 0 {
			continue
		}

		barSlotWidth := (plotWidth / math.Max(float64(seriesWithData), 1)) * 0.75
		maxBarWidth := 28.0
		if barSlotWidth > maxBarWidth {
			barSlotWidth = maxBarWidth
		}

		for groupIdx, group := range groups {
			xOffset := (float64(groupIdx) - float64(len(groups)-1)/2.0) * barSlotWidth
			stackedBase := map[float64]float64{}
			stackedNegativeBase := map[float64]float64{}

			for _, series := range group.Series {
				for _, point := range series.Data {
					x := plotX + (point.X-xMin)/(xMax-xMin)*plotWidth + xOffset
					barTopValue := point.Y
					barBottomValue := 0.0

					if strings.HasPrefix(group.Key, "stack:") {
						if point.Y >= 0 {
							barBottomValue = stackedBase[point.X]
							barTopValue = barBottomValue + point.Y
							stackedBase[point.X] = barTopValue
						} else {
							barBottomValue = stackedNegativeBase[point.X]
							barTopValue = barBottomValue + point.Y
							stackedNegativeBase[point.X] = barTopValue
						}
					}

					yTop := plotY + plotHeight - (barTopValue-yMin)/(yMax-yMin)*plotHeight
					yBottom := plotY + plotHeight - (barBottomValue-yMin)/(yMax-yMin)*plotHeight
					height := yBottom - yTop
					if height < 0 {
						yTop, height = yBottom, -height
					}

					svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" opacity="0.9"/>`,
						x-barSlotWidth/2, yTop, barSlotWidth, height, series.Color))
				}
			}
		}
	}

	for _, series := range c.Series {
		if len(series.Data) == 0 {
			continue
		}
		if series.Style == "bar" {
			continue
		}

		// Get the scale for this series
		scale, ok := scaleMap[series.YAxis]
		if !ok {
			// Fallback to first scale if series scale not found
			if len(c.Scales) > 0 {
				scale = &c.Scales[0]
			} else {
				continue
			}
		}

		yMin := scale.Min
		yMax := scale.Max

		if series.Style == "marker" {
			for _, point := range series.Data {
				x := plotX + (point.X-xMin)/(xMax-xMin)*plotWidth
				y := plotY + plotHeight - (point.Y-yMin)/(yMax-yMin)*plotHeight
				svg.WriteString(fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="3" fill="%s"/>`, x, y, series.Color))
				label := fmt.Sprintf("%.0f", point.Y)
				if point.YIsDuration {
					label = point.YLabel
				}
				svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" text-anchor="middle" font-size="10" fill="%s">%s</text>`,
					x, y-8, series.Color, label))
			}
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

func (c *Chart) generateLegend(layout legendLayout) string {
	var svg strings.Builder

	legendX, legendY := layout.X, layout.Y
	legendWidth, legendHeight := layout.Width, layout.Height

	svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="white" stroke="#ccc" stroke-width="1" rx="3"/>`,
		legendX, legendY, legendWidth, legendHeight))

	const (
		padding   = 8.0
		rowHeight = 20.0
		lineWidth = 15.0
		textGap   = 6.0
		itemGap   = 14.0
	)

	if layout.Orientation == "left" || layout.Orientation == "right" {
		for i, series := range c.Series {
			y := legendY + padding + 12 + float64(i)*rowHeight
			svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="3" fill="%s"/>`,
				legendX+padding, y-2, lineWidth, series.Color))
			svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" font-size="11" fill="#333">%s</text>`,
				legendX+padding+lineWidth+textGap, y+2, series.Name))
		}
		return svg.String()
	}

	// Horizontal legend (top/bottom)
	currentX := legendX + padding
	baselineY := legendY + padding + 12
	for i, series := range c.Series {
		svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="3" fill="%s"/>`,
			currentX, baselineY-2, lineWidth, series.Color))
		svg.WriteString(fmt.Sprintf(`<text x="%.2f" y="%.2f" font-size="11" fill="#333">%s</text>`,
			currentX+lineWidth+textGap, baselineY+2, series.Name))

		textWidth := float64(len([]rune(series.Name))) * 11.0 * 0.6
		itemWidth := lineWidth + textGap + textWidth
		if i < len(c.Series)-1 {
			itemWidth += itemGap
		}
		currentX += itemWidth
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
