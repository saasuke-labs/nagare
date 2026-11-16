package chart

// The watercolor look for charts, built from the same painted primitives as
// diagrams: a paper sheet, ink axes, pencil grid, pen lines in each series'
// pigment with a faint wash under them, and bars laid as washes.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/diagram/components/style"
)

func (c *Chart) painted() bool { return c.Look == style.LookWatercolor }

func (c *Chart) theme() style.Theme { return style.Light.Watercolor() }

// textColor and labelColor are the two text colours; the flat look keeps its
// historical greys.
func (c *Chart) textColor() string {
	if c.painted() {
		return c.theme().Text
	}
	return "#333"
}

func (c *Chart) labelColor() string {
	if c.painted() {
		return c.theme().Muted
	}
	return "#666"
}

func (c *Chart) seed(parts ...string) uint64 {
	return style.SeedFor(append([]string{"chart", c.Title}, parts...)...)
}

// paintedOpen writes the root element, typography, filters and the sheet.
func (c *Chart) paintedOpen() string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg" class="nagare-watercolor">`, c.Width, c.Height, c.Width, c.Height)
	b.WriteString(style.WatercolorCSS())
	b.WriteString(style.WatercolorDefs())
	if !c.Paperless {
		b.WriteString(style.Paper(c.Width, c.Height, c.theme()))
	}
	b.WriteString(`<g style="isolation:isolate">`)
	return b.String()
}

const paintedClose = `</g></svg>`

// pencil draws a faint hand-ruled line (grid).
func (c *Chart) pencil(x1, y1, x2, y2 float64, key string) string {
	t := c.theme()
	return style.InkStroke([]style.Pt{{X: x1, Y: y1}, {X: x2, Y: y2}}, style.InkSpec{
		Width: 0.75, Color: t.Muted, Seed: c.seed("grid", key), Tremor: 0.5, TaperIn: 20, TaperOut: 20, Opacity: 0.32,
	})
}

// axisLine draws an axis in ink.
func (c *Chart) axisLine(x1, y1, x2, y2 float64, key string) string {
	t := c.theme()
	return `<g filter="url(#nagare-wc-ink)">` + style.InkStroke([]style.Pt{{X: x1, Y: y1}, {X: x2, Y: y2}}, style.InkSpec{
		Width: 1.6, Color: t.Ink, Seed: c.seed("axis", key), Tremor: 0.6, TaperIn: 6, TaperOut: 14,
	}) + `</g>`
}

// seriesPigment is the paint for a series colour.
func seriesPigment(color string) string { return style.Pigment(color) }

// paintBar lays one bar as a wash, settled toward its base.
func (c *Chart) paintBar(x, y, w, h float64, color, key string) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	poly := style.RoundedRect(x, y, w, h, style.Clamp(w*0.12, 0.5, 3), 12)
	return style.Wash(style.WashSpec{
		Poly: poly, Pigment: seriesPigment(color), Density: 0.5, Seed: c.seed("bar", key),
		Layers: 4, Wander: 0.06, MaxOff: style.Clamp(w*0.08, 0.6, 2.2), Settle: 0.5,
	}, c.theme())
}

// paintDot is a small dot of the series' ink for a data point.
func (c *Chart) paintDot(x, y, r float64, color, key string) string {
	_ = key
	return fmt.Sprintf(`<circle cx="%s" cy="%s" r="%s" fill="%s" fill-opacity="0.9" filter="url(#nagare-wc-ink)"/>`,
		style.F(x), style.F(y), style.F(r), style.Shade(seriesPigment(color), 0.15))
}

// paintLine draws a series as a pen line in its pigment. A solid line also
// gets a faint wash down to the baseline, the way a painter would tint the
// area under a curve; dashed and dotted lines are runs of short strokes.
func (c *Chart) paintLine(pts []style.Pt, baseline float64, color, lineStyle, key string) string {
	if len(pts) == 0 {
		return ""
	}
	pigment := seriesPigment(color)
	t := c.theme()
	var b strings.Builder
	if lineStyle != "dashed" && lineStyle != "dotted" && len(pts) > 1 {
		// A graded wash: bands that hang from the line to a fraction of the
		// way down, so the colour is richest at the line and fades out.
		for i, depth := range []float64{1, 0.45} {
			band := append([]style.Pt{}, pts...)
			for j := len(pts) - 1; j >= 0; j-- {
				p := pts[j]
				band = append(band, style.Pt{X: p.X, Y: p.Y + (baseline-p.Y)*depth})
			}
			b.WriteString(style.Wash(style.WashSpec{
				Poly: band, Pigment: pigment, Density: 0.09, Seed: c.seed("area", key, strconv.Itoa(i)), Layers: 2, Wander: 0.04, MaxOff: 3,
				Filter: "nagare-wc-soft",
			}, t))
		}
	}
	spec := style.InkSpec{Width: 2.1, Color: style.Shade(pigment, 0.08), Seed: c.seed("line", key), Tremor: 0.5}
	b.WriteString(`<g filter="url(#nagare-wc-ink)">`)
	switch lineStyle {
	case "dashed":
		b.WriteString(style.InkDashes(pts, spec, 10, 6))
	case "dotted":
		spec.Width = 2.3
		b.WriteString(style.InkDashes(pts, spec, 2.2, 4.5))
	default:
		spec.TaperIn, spec.TaperOut = 6, 6
		b.WriteString(style.InkStroke(pts, spec))
	}
	b.WriteString(`</g>`)
	return b.String()
}

// paintLegendBox draws the legend frame as a light pen loop on bare paper.
func (c *Chart) paintLegendBox(x, y, w, h float64) string {
	t := c.theme()
	poly := style.RoundedRect(x, y, w, h, 4, 10)
	return `<g filter="url(#nagare-wc-ink)">` + style.InkLoop(poly, style.InkSpec{Width: 0.9, Color: t.Muted, Seed: c.seed("legend"), Opacity: 0.8, NoLift: true}) + `</g>`
}

// paintSwatch is a short brush dab in the series' pigment.
func (c *Chart) paintSwatch(x, y, w float64, color, lineStyle string, i int) string {
	key := "swatch" + strconv.Itoa(i)
	if lineStyle == "bar" || lineStyle == "marker" {
		return c.paintBar(x, y-3, w, 7, color, key)
	}
	pts := []style.Pt{{X: x, Y: y + 0.5}, {X: x + w, Y: y + 0.5}}
	return c.paintLine(pts, y, color, lineStyle, key)
}
