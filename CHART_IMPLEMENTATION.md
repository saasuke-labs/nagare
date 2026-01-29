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
9. **chart-multi-scale.nagare** - Multi-scale example with duration and distance (NEW)

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
- Tooltips/hover information

## Multi-Scale Support (NEW)

Charts now support multiple Y-axis scales, allowing you to plot different metrics with different units on the same chart. This prevents the "almost flat" rendering issue when mixing values with vastly different ranges (e.g., duration in seconds vs distance in kilometers).

### Basic Multi-Scale Usage

Define separate scales and assign each series to a scale using the `yaxis` property:

```
chart
title: Training Metrics - Duration vs Distance
xaxis: date

scale
  id: duration
  label: Duration (minutes)
  type: duration

scale
  id: distance
  label: Distance (km)

series: total-duration
yaxis: duration
type: duration
data:
  2025-01-01: 60:00
  2025-01-02: 62:30

series: jog-distance
yaxis: distance
data:
  2025-01-01: 5.0
  2025-01-02: 5.2
```

**Result:**
- Left Y-axis shows duration scale (for total-duration)
- Right Y-axis shows distance scale (for jog-distance)
- Each series is properly scaled to its own axis

### Scale Properties

Each scale block supports:
- **id** (required) - Unique identifier for the scale
- **label** (optional) - Axis label text
- **type** (optional) - `number` (default) or `duration`
- **min** (optional) - Manual minimum value (auto-calculated if not specified)
- **max** (optional) - Manual maximum value (auto-calculated if not specified)

### Manual Range Specification

You can manually set the range for a scale:

```
scale
  id: percentage
  label: Success Rate (%)
  min: 0
  max: 100
```

When min/max are specified, auto-calculation is disabled for that scale.

### Backward Compatibility

Old charts without scale definitions continue to work. A default scale is automatically created:
- Always initialized with type "number"
- Uses `ylabel` (if present) for the scale label
- All series are assigned to the "default" scale
- Range is auto-calculated based on all data points

### Multiple Scales Rendering

- **First scale** - Rendered on the left Y-axis
- **Second scale** - Rendered on the right Y-axis
- Additional scales beyond two are currently supported in the data model but only the first two are rendered

### Use Cases

Perfect for charts that combine:
- Duration and distance (running/training metrics)
- Time and count (performance monitoring)
- Temperature and humidity (weather data)
- Price and volume (financial charts)
- Any metrics with different units or value ranges
