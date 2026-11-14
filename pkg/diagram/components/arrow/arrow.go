package arrow

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	"sync/atomic"
)

//go:embed arrow.html
var templateFile embed.FS

var tmpl = template.Must(template.New("arrow").Funcs(template.FuncMap{
	"add": func(a, b float64) float64 { return a + b },
	"mul": func(a, b float64) float64 { return a * b },
	"sub": func(a, b float64) float64 { return a - b },
}).ParseFS(templateFile, "arrow.html"))

type Point struct {
	X float64
	Y float64
}

type Component struct {
	Points      []Point
	StrokeColor string
	StrokeWidth float64
	Style       string
	MarkerStart bool
	MarkerEnd   bool
	markerID    string
}

var markerCounter uint64

func ResetMarkerCounter() {
	atomic.StoreUint64(&markerCounter, 0)
}

func nextMarkerID() string {
	id := atomic.AddUint64(&markerCounter, 1)
	return fmt.Sprintf("arrowhead-%d", id)
}

func New(points []Point) *Component {
	return &Component{Points: points, StrokeColor: "#1f2937", StrokeWidth: 2, MarkerEnd: true, markerID: nextMarkerID()}
}

func (a *Component) ensureMarkerID() string {
	if a.markerID == "" {
		a.markerID = nextMarkerID()
	}
	return a.markerID
}

func (a *Component) Draw() string {
	if len(a.Points) < 2 {
		return ""
	}

	trimmedStyle := strings.TrimSpace(a.Style)

	data := struct {
		Points      []Point
		StrokeColor string
		StrokeWidth float64
		Style       string
		HasStyle    bool
		IsDashed    bool
		MarkerStart bool
		MarkerEnd   bool
		MarkerID    string
	}{
		Points:      a.Points,
		StrokeColor: a.StrokeColor,
		StrokeWidth: a.StrokeWidth,
		Style:       trimmedStyle,
		HasStyle:    trimmedStyle != "",
		IsDashed:    trimmedStyle == "dashed",
		MarkerStart: a.MarkerStart,
		MarkerEnd:   a.MarkerEnd,
		MarkerID:    a.ensureMarkerID(),
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "arrow", data); err != nil {
		return fmt.Sprintf("<!-- Error rendering arrow template: %v -->", err)
	}
	return buf.String()
}
