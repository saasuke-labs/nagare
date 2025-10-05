package components

import (
	"fmt"

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

	result, err := RenderTemplate("rectangle", data)
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering rectangle template: %v -->", err)
	}

	return result
}
