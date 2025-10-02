package components

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/saasuke-labs/nagare/pkg/props"
)

type Component interface {
	Draw() string
}

type Shape struct {
	Width         float64
	Height        float64
	X             float64
	Y             float64
	AlignmentRefs map[string]string // Store alignment references for later resolution
}

type RectangleProps struct {
	Title           string `prop:"title"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
}

func (r *RectangleProps) Parse(input string) error {
	return props.ParseProps(input, r)
}

func DefaultRectangleProps() RectangleProps {
	return RectangleProps{
		Title:           "",
		BackgroundColor: "#e6f3ff",
		ForegroundColor: "#333333",
	}
}

type Rectangle struct {
	Shape
	Text  string
	Props RectangleProps
	State string
}

func NewRectangle(id string) *Rectangle {
	return &Rectangle{
		Text:  id,
		Props: DefaultRectangleProps(),
	}
}

func (r *Rectangle) Draw() string {
	displayText := r.Text
	if r.Props.Title != "" {
		displayText = r.Props.Title
	}

	data := struct {
		X             float64
		Y             float64
		Width         float64
		Height        float64
		DisplayText   string
		Background    string
		Foreground    string
		BorderRadiusX float64
		BorderRadiusY float64
	}{
		X:             r.X,
		Y:             r.Y,
		Width:         r.Width,
		Height:        r.Height,
		DisplayText:   displayText,
		Background:    r.Props.BackgroundColor,
		Foreground:    r.Props.ForegroundColor,
		BorderRadiusX: r.Height * 0.1,
		BorderRadiusY: r.Height * 0.1,
	}

	const rectangleTemplate = `<g transform="translate({{printf "%.6f" .X}},{{printf "%.6f" .Y}})">
    <rect x="0" y="0" width="{{printf "%.6f" .Width}}" height="{{printf "%.6f" .Height}}"
          rx="{{printf "%.6f" .BorderRadiusX}}" ry="{{printf "%.6f" .BorderRadiusY}}"
          fill="{{.Background}}" stroke="{{.Foreground}}" stroke-width="2"/>
    <text x="{{printf "%.6f" (mul .Width 0.5)}}" y="{{printf "%.6f" (mul .Height 0.5)}}"
          font-family="Arial" font-size="14" fill="{{.Foreground}}"
          text-anchor="middle" dominant-baseline="middle">{{.DisplayText}}</text>
</g>`

	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
	}

	tmpl, err := template.New("rectangle").Funcs(funcMap).Parse(rectangleTemplate)
	if err != nil {
		return fmt.Sprintf("<!-- Error parsing rectangle template: %v -->", err)
	}

	var svg bytes.Buffer
	if err := tmpl.Execute(&svg, data); err != nil {
		return fmt.Sprintf("<!-- Error executing rectangle template: %v -->", err)
	}

	return svg.String()
}
