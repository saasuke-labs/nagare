package components

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Point struct {
	X float64
	Y float64
}

type Arrow struct {
	Points      []Point
	StrokeColor string
	StrokeWidth float64
	Style       string
	MarkerStart bool
	MarkerEnd   bool
	markerID    string
}

var arrowMarkerCounter uint64

func nextArrowMarkerID() string {
	id := atomic.AddUint64(&arrowMarkerCounter, 1)
	return fmt.Sprintf("arrowhead-%d", id)
}

func NewArrow(points []Point) *Arrow {
	markerID := nextArrowMarkerID()
	return &Arrow{
		Points:      points,
		StrokeColor: "#1f2937",
		StrokeWidth: 2,
		MarkerEnd:   true,
		markerID:    markerID,
	}
}

func (a *Arrow) ensureMarkerID() string {
	if a.markerID == "" {
		a.markerID = nextArrowMarkerID()
	}
	return a.markerID
}

func (a *Arrow) Draw() string {
	if len(a.Points) < 2 {
		return ""
	}

	markerID := a.ensureMarkerID()

	var styleAttr string
	if strings.TrimSpace(a.Style) != "" {
		styleAttr = fmt.Sprintf(" style=\"%s\"", strings.TrimSpace(a.Style))
	} else {
		styleAttr = fmt.Sprintf(" stroke=\"%s\" stroke-width=\"%.2f\" fill=\"none\" stroke-linecap=\"round\" stroke-linejoin=\"round\"", a.StrokeColor, a.StrokeWidth)
	}

	markerAttributes := ""
	defs := ""
	if a.MarkerStart || a.MarkerEnd {
		defs = fmt.Sprintf(`<defs><marker id="%s" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="%s"/></marker></defs>`, markerID, a.StrokeColor)
		if a.MarkerStart {
			markerAttributes += fmt.Sprintf(" marker-start=\"url(#%s)\"", markerID)
		}
		if a.MarkerEnd {
			markerAttributes += fmt.Sprintf(" marker-end=\"url(#%s)\"", markerID)
		}
	}

	coords := make([]string, 0, len(a.Points))
	for _, pt := range a.Points {
		coords = append(coords, fmt.Sprintf("%.2f,%.2f", pt.X, pt.Y))
	}

	return fmt.Sprintf(`<g class="connector">%s<polyline points="%s"%s%s/></g>`,
		defs,
		strings.Join(coords, " "),
		styleAttr,
		markerAttributes,
	)
}
