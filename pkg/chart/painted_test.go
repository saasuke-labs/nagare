package chart

import (
	"strings"
	"testing"
)

const paintedChartSrc = `chart
title: Deploys
style: watercolor
xaxis: number
legend: top

series: api | web | target | trend
color: #3b82f6 | #f59e0b | #ec4899 | #14b8a6
style: bar | bar | marker | dashed
stack: main | main | none | none
data:
  1: 12 | 5 | 20 | 15
  2: 15 | 7 | 20 | 18
  3: 9 | 11 | 22 | 21
`

func TestPaintedChart(t *testing.T) {
	c, err := Parse(paintedChartSrc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if c.Look != "watercolor" {
		t.Fatalf("expected style: watercolor to set the chart's look, got %q", c.Look)
	}
	for _, s := range c.Series {
		if s.Style == "watercolor" {
			t.Fatal("style: watercolor must not become a series style")
		}
	}
	if got := []string{c.Series[0].Style, c.Series[2].Style, c.Series[3].Style}; got[0] != "bar" || got[1] != "marker" || got[2] != "dashed" {
		t.Fatalf("series styles changed: %v", got)
	}

	svg := c.RenderSVG()
	if svg != mustRender(t, paintedChartSrc) {
		t.Fatal("expected a painted chart to render identically every time")
	}
	for _, want := range []string{`class="nagare-watercolor"`, `class="nagare-paper"`, `class="nagare-wash`, `filter="url(#nagare-wc-ink)"`} {
		if !strings.Contains(svg, want) {
			t.Errorf("expected painted chart to contain %q", want)
		}
	}
	for _, flat := range []string{`fill="white"`, `<line `, `fill="#333"`, `fill="#666"`, `stroke-dasharray`} {
		if strings.Contains(svg, flat) {
			t.Errorf("painted chart still contains flat %q", flat)
		}
	}
}

func TestPaintedChartPaperCanBeLeftOut(t *testing.T) {
	svg := mustRender(t, strings.Replace(paintedChartSrc, "style: watercolor\n", "style: watercolor\npaper: none\n", 1))
	if strings.Contains(svg, `class="nagare-paper"`) {
		t.Fatal("expected paper: none to leave the canvas transparent")
	}
}

func TestFlatChartIsNotPainted(t *testing.T) {
	svg := mustRender(t, strings.Replace(paintedChartSrc, "style: watercolor\n", "", 1))
	if strings.Contains(svg, "nagare-") || !strings.Contains(svg, `fill="white"`) {
		t.Fatal("expected the flat chart without style: watercolor")
	}
}

func mustRender(t *testing.T, src string) string {
	t.Helper()
	c, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return c.RenderSVG()
}
