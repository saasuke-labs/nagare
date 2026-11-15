package chart

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// yAxisIntervals is the number of intervals on every Y axis. Keeping it the
// same for all scales means the left axis, the right axis and the horizontal
// grid lines always line up.
const yAxisIntervals = 5

// minXTickSpacing is the smallest distance, in pixels, between two X labels.
const minXTickSpacing = 48.0

const secondsPerDay = 86400.0

// axisTick is one labelled position on the X axis.
type axisTick struct {
	Pos   float64 // pixel position along the axis
	Label string
}

// niceStep returns the smallest round step (1, 2, 2.5 or 5 times a power of
// ten) that is not smaller than raw.
func niceStep(raw float64) float64 {
	if raw <= 0 || math.IsInf(raw, 0) || math.IsNaN(raw) {
		return 1
	}
	magnitude := math.Pow(10, math.Floor(math.Log10(raw)))
	for _, factor := range []float64{1, 2, 2.5, 5, 10} {
		step := factor * magnitude
		if step >= raw*(1-1e-9) {
			return step
		}
	}
	return 10 * magnitude
}

// durationSteps are the steps, in seconds, that read naturally on a time axis.
var durationSteps = []float64{
	1, 2, 5, 10, 15, 30, // seconds
	60, 120, 300, 600, 900, 1800, // minutes
	3600, 7200, 10800, 21600, 43200, 86400, // hours
}

// niceDurationStep is niceStep for values measured in seconds.
func niceDurationStep(raw float64) float64 {
	for _, step := range durationSteps {
		if step >= raw*(1-1e-9) {
			return step
		}
	}
	return niceStep(raw/secondsPerDay) * secondsPerDay
}

// niceBounds widens [lo, hi] to bounds that divide into yAxisIntervals round
// steps, so every tick on the axis is a round value.
func niceBounds(lo, hi float64, duration bool) (float64, float64) {
	if hi <= lo {
		hi = lo + 1
	}

	pick := niceStep
	if duration {
		pick = niceDurationStep
	}

	step := pick((hi - lo) / yAxisIntervals)
	for i := 0; i < 64; i++ {
		min := math.Floor(lo/step+1e-9) * step
		max := min + yAxisIntervals*step
		if max >= hi-step*1e-9 {
			return min, max
		}
		// Snapping the lower bound down left the top uncovered: try the next step.
		step = pick(step * 1.000001)
	}
	return lo, hi
}

// formatTick formats an axis value with only as many decimals as its step
// needs: 20 rather than 20.0, but 0.25 when the step is 0.25.
func formatTick(value, step float64) string {
	return formatNumber(value, stepDecimals(step))
}

func stepDecimals(step float64) int {
	step = math.Abs(step)
	if step == 0 {
		return 0
	}
	for decimals := 0; decimals <= 4; decimals++ {
		scaled := step * math.Pow(10, float64(decimals))
		if math.Abs(scaled-math.Round(scaled)) < 1e-6*math.Max(1, scaled) {
			return decimals
		}
	}
	return 2
}

func formatNumber(value float64, decimals int) string {
	label := strconv.FormatFloat(value, 'f', decimals, 64)
	if strings.Contains(label, ".") {
		label = strings.TrimRight(strings.TrimRight(label, "0"), ".")
	}
	if label == "-0" {
		return "0"
	}
	return label
}

// distinctXValues returns the sorted X values used by any series.
func (c *Chart) distinctXValues() []float64 {
	seen := map[float64]struct{}{}
	for _, series := range c.Series {
		for _, point := range series.Data {
			seen[point.X] = struct{}{}
		}
	}
	values := make([]float64, 0, len(seen))
	for x := range seen {
		values = append(values, x)
	}
	sort.Float64s(values)
	return values
}

// xTicks chooses the labelled positions of the X axis.
//
// When there are few enough data points to label each one, every point gets
// its own tick. Otherwise ticks fall on round values (numbers) or calendar
// boundaries (dates). Either way a label sits exactly where its value is.
func (c *Chart) xTicks(plotX, plotWidth, xMin, xMax float64) []axisTick {
	values := c.distinctXValues()
	if len(values) == 0 || xMax <= xMin || plotWidth <= 0 {
		return nil
	}

	maxTicks := int(plotWidth / minXTickSpacing)
	if maxTicks < 2 {
		maxTicks = 2
	}

	var tickValues []float64
	var label func(float64) string

	if c.XAxisType == "date" {
		tickValues, label = dateTicks(values, maxTicks)
	} else {
		tickValues, label = numberTicks(values, maxTicks)
	}

	ticks := make([]axisTick, 0, len(tickValues))
	lastPos := math.Inf(-1)
	for _, value := range tickValues {
		pos := plotX + (value-xMin)/(xMax-xMin)*plotWidth
		// Unevenly spaced data can put two labels on top of each other.
		if pos-lastPos < minXTickSpacing*0.6 {
			continue
		}
		ticks = append(ticks, axisTick{Pos: pos, Label: label(value)})
		lastPos = pos
	}
	return ticks
}

func numberTicks(values []float64, maxTicks int) ([]float64, func(float64) string) {
	if len(values) <= maxTicks {
		decimals := 0
		for _, value := range values {
			if d := stepDecimals(value); d > decimals {
				decimals = d
			}
		}
		return values, func(value float64) string { return formatNumber(value, decimals) }
	}

	first, last := values[0], values[len(values)-1]
	step := niceStep((last - first) / float64(maxTicks-1))
	var ticks []float64
	for i := 0; i < 64; i++ {
		ticks = ticks[:0]
		start := math.Ceil(first/step-1e-9) * step
		for value := start; value <= last+step*1e-9; value += step {
			ticks = append(ticks, value)
		}
		if len(ticks) <= maxTicks {
			break
		}
		step = niceStep(step * 1.000001)
	}
	return ticks, func(value float64) string { return formatTick(value, step) }
}

func dateTicks(values []float64, maxTicks int) ([]float64, func(float64) string) {
	first := time.Unix(int64(values[0]), 0).UTC()
	last := time.Unix(int64(values[len(values)-1]), 0).UTC()
	multiYear := first.Year() != last.Year()

	dayLabel := func(value float64) string {
		return time.Unix(int64(value), 0).UTC().Format("01/02")
	}
	// On an axis that spans several years, January carries the year.
	monthLabel := func(value float64) string {
		t := time.Unix(int64(value), 0).UTC()
		if multiYear && t.Month() == time.January {
			return t.Format("2006")
		}
		return t.Format("Jan")
	}
	yearLabel := func(value float64) string {
		return time.Unix(int64(value), 0).UTC().Format("2006")
	}

	// Few enough points to label each of them.
	if len(values) <= maxTicks {
		if isMonthlySeries(values) {
			return values, monthLabel
		}
		return values, dayLabel
	}

	// Whole days, counted from the first data point.
	spanDays := (values[len(values)-1] - values[0]) / secondsPerDay
	for _, stepDays := range []float64{1, 2, 7, 14} {
		if int(spanDays/stepDays)+1 > maxTicks {
			continue
		}
		var ticks []float64
		for value := values[0]; value <= values[len(values)-1]+1; value += stepDays * secondsPerDay {
			ticks = append(ticks, value)
		}
		return ticks, dayLabel
	}

	// First day of every month, every other month, quarter or half year.
	for _, stepMonths := range []int{1, 2, 3, 6} {
		ticks := monthStarts(first, last, stepMonths)
		if len(ticks) >= 2 && len(ticks) <= maxTicks {
			return ticks, monthLabel
		}
	}

	// First of January, every one or more years.
	years := float64(last.Year() - first.Year())
	stepYears := int(math.Max(1, math.Ceil(niceStep(years/float64(maxTicks-1)))))
	var ticks []float64
	for year := first.Year(); year <= last.Year()+1; year++ {
		if year%stepYears != 0 {
			continue
		}
		t := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		if t.Before(first) || t.After(last) {
			continue
		}
		ticks = append(ticks, float64(t.Unix()))
	}
	return ticks, yearLabel
}

// isMonthlySeries reports whether the points are one calendar month apart.
func isMonthlySeries(values []float64) bool {
	if len(values) < 2 {
		return false
	}
	for i := 1; i < len(values); i++ {
		gapDays := (values[i] - values[i-1]) / secondsPerDay
		if gapDays < 28 || gapDays > 31 {
			return false
		}
	}
	return true
}

// monthStarts returns the first day of every stepMonths-th month in
// [first, last], aligned so that January is always one of them.
func monthStarts(first, last time.Time, stepMonths int) []float64 {
	var ticks []float64
	t := time.Date(first.Year(), first.Month(), 1, 0, 0, 0, 0, time.UTC)
	for ; !t.After(last); t = t.AddDate(0, 1, 0) {
		if t.Before(first) || (int(t.Month())-1)%stepMonths != 0 {
			continue
		}
		ticks = append(ticks, float64(t.Unix()))
	}
	return ticks
}
