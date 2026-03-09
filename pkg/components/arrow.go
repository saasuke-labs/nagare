package components

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/saasuke-labs/nagare/pkg/diagram/components/core"
)

// Point is the canonical 2D coordinate type, aliased from core.
type Point = core.Point

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

// ResetArrowMarkerCounter resets the global arrow marker counter.
// This is primarily useful for testing to ensure consistent IDs.
func ResetArrowMarkerCounter() {
	atomic.StoreUint64(&arrowMarkerCounter, 0)
}

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

	trimmedStyle := strings.TrimSpace(a.Style)

	data := struct {
		Points      []Point
		StrokeColor string
		StrokeWidth float64
		Style       string
		HasStyle    bool
		MarkerStart bool
		MarkerEnd   bool
		MarkerID    string
	}{
		Points:      a.Points,
		StrokeColor: a.StrokeColor,
		StrokeWidth: a.StrokeWidth,
		Style:       trimmedStyle,
		HasStyle:    trimmedStyle != "",
		MarkerStart: a.MarkerStart,
		MarkerEnd:   a.MarkerEnd,
		MarkerID:    markerID,
	}

	result, err := RenderTemplate("arrow", data)
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering arrow template: %v -->", err)
	}

	return result
}
