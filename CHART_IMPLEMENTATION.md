# Chart Implementation Summary

## Overview

Implemented a completely separate chart system for nagare with its own parser, independent from the diagram DSL.

## Key Features

### Simplified Input Format

- **Separate `chart` identifier** at the start - diagrams and charts are completely independent
- **Multiline data support** - much more readable for large datasets
- **Clean key-value structure** - no complex nesting or parentheses
- **Dual-format support** - both classic multi-block and compact pipe-delimited formats

### Format 1: Classic Multi-Block (One Series Per Block)

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

### Format 2: Compact Pipe-Delimited (Multiple Series in One Block)

For charts with multiple series, use pipe delimiters to save space:

```
chart
title: Training Metrics (Compact)
xaxis: date

series: distance | pace
color: #3b82f6 | #ef4444
style: line | dashed
data:
  2025-01-01: 3.01 | 6.5
  2025-01-02: 3.5 | 6.3
  2025-01-03: 4.2 | 6.1
```

Both formats are fully supported and can be mixed in the same document.

### Duration Type Support

Charts now support duration values in addition to numbers. Use the `type` field to specify whether data should be parsed as durations:

```
chart
title: Running Times
xaxis: date

series: 5K Time
type: duration
data:
  2025-01-01: 25:30
  2025-01-08: 24:45
  2025-01-15: 24:15
```

**Duration Formats:**
- **MM:SS** - Minutes and seconds (e.g., `5:30` = 5 minutes, 30 seconds)
- **HH:MM:SS** - Hours, minutes, and seconds (e.g., `3:02:31` = 3 hours, 2 minutes, 31 seconds)

Durations are converted to seconds internally for calculations. Y-axis labels show the scale range in duration format (interpolated values from min to max), not the original data point values.

**Mixed Types:**

You can mix duration and number types in the same chart using pipe-delimited format:

```
chart
title: Training Metrics

series: Lap Time | Rest Time | Distance
type: duration | duration | number
data:
  2025-01-01: 5:30 | 2:15 | 42.5
  2025-01-02: 5:15 | 2:00 | 45.2
```

**Note:** When mixing duration and number types in the same chart, all Y-axis labels will be formatted as durations if any series has type "duration". For best results, use separate charts for duration and number data, or ensure all values are of the same type.

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
6. **chart-duration-running.nagare** - Duration type example (5K and 10K running times)
7. **chart-duration-mixed.nagare** - Mixed duration and number types (lap time, rest time, distance)
8. **chart-duration-marathon.nagare** - Long duration format (marathon times in HH:MM:SS)

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
- **Duration Y-values** - Support for time durations in MM:SS or HH:MM:SS format
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
