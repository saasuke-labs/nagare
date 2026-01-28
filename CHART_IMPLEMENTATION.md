# Chart Implementation Summary

## Overview

Implemented a completely separate chart system for nagare with its own parser, independent from the diagram DSL.

## Key Features

### Simplified Input Format

- **Separate `chart` identifier** at the start - diagrams and charts are completely independent
- **Multiline data support** - much more readable for large datasets
- **Clean key-value structure** - no complex nesting or parentheses

### Example Format

```
chart
title: Training Metrics
width: 900
height: 600
xaxis: date
xlabel: Date
ylabel: Distance (km)
legend: top-right
grid: true

series: distance
color: #3b82f6
style: line
data:
  2025-01-01: 3.01
  2025-01-02: 3.5
  2025-01-03: 4.2

series: pace
color: #ef4444
style: dashed
data:
  2025-01-01: 6.5
  2025-01-02: 6.3
  2025-01-03: 6.1
```

## Architecture

### New Package: `pkg/chart/`

- **chart.go** - Complete chart parser and SVG rendering engine
- **chart_test.go** - Comprehensive test suite

### Files Modified

- **cmd/main.go** - Added routing logic to detect "chart" prefix and delegate to appropriate parser
- **pkg/layout/layout.go** - Removed LineChart references (no longer needed in layout system)

### Test Diagrams

Located in `.github/testdiagrams/`:

1. **chart-single-line.nagare** - Single series with date axis (running distance)
2. **chart-multi-series.nagare** - Multiple series with different styles (training metrics)
3. **chart-temperature.nagare** - Three series, dotted line example (temperature monitoring)
4. **chart-sales.nagare** - Multi-line sales data with legend positioning (monthly sales)
5. **chart-network.nagare** - Single series, no legend (network latency)

All diagrams use the **new multiline format** for data entry.

## Advantages

✅ **Cleaner syntax** - No comma-separated pairs, indentation shows data blocks  
✅ **More readable** - Each data point on its own line  
✅ **No diagram pollution** - Complete separation from diagram parser  
✅ **Extensible** - Easy to add new chart types (bar, pie, scatter, etc.)  
✅ **Type-safe routing** - Client code checks first line to route to correct parser

## Supported Features

**Data:**

- Multiple series per chart
- Date and numeric X-axis types
- Auto-scaling with padding

**Styling:**

- Customizable colors
- Line styles: solid, dashed, dotted
- Grid lines (toggle)
- Legend positioning: top-left, top-right, bottom-left, bottom-right, none

**Labels:**

- Chart title
- X and Y axis labels
- Series names in legend

## Testing

All tests pass:

```
=== RUN   TestParseSimpleChart
--- PASS: TestParseSimpleChart (0.00s)
=== RUN   TestParseMultiSeries
--- PASS: TestParseMultiSeries (0.00s)
=== RUN   TestRenderSVG
--- PASS: TestRenderSVG (0.00s)
PASS
```

## Future Enhancements

- Bar charts
- Scatter plots
- Pie charts
- Stacked area charts
- Custom axis ranges
- Multiple Y-axes
- Tooltips/hover information
