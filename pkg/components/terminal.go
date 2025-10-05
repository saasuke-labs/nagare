package components

import (
	"fmt"

	"github.com/saasuke-labs/nagare/pkg/props"
)

// TerminalProps defines configurable properties for a Terminal component.
type TerminalProps struct {
	Title           string `prop:"title"`
	WorkingDir      string `prop:"cwd"`
	Command         string `prop:"command"`
	PromptSymbol    string `prop:"prompt"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
	AccentColor     string `prop:"accent"`
}

// Parse implements the propertyParser interface.
func (t *TerminalProps) Parse(input string) error {
	return props.ParseProps(input, t)
}

// DefaultTerminalProps provides sensible defaults for a terminal window.
func DefaultTerminalProps() TerminalProps {
	return TerminalProps{
		Title:           "Terminal",
		WorkingDir:      "~/project",
		Command:         "go run main.go",
		PromptSymbol:    "$",
		BackgroundColor: "#0f172a",
		ForegroundColor: "#f8fafc",
		AccentColor:     "#38bdf8",
	}
}

// Terminal renders a faux terminal window with a command prompt.
type Terminal struct {
	Shape
	Text  string
	Props TerminalProps
	State string
}

// NewTerminal creates a terminal instance with default props.
func NewTerminal(id string) *Terminal {
	return &Terminal{
		Shape: Shape{},
		Text:  id,
		Props: DefaultTerminalProps(),
	}
}

type TerminalTemplateData struct {
	X              float64
	Y              float64
	Width          float64
	Height         float64
	Props          TerminalProps
	Text           string
	HeaderControls HeaderControlProps
}

func (t *Terminal) templateData() TerminalTemplateData {
	return TerminalTemplateData{
		X:              t.X,
		Y:              t.Y,
		Width:          t.Width,
		Height:         t.Height,
		Props:          t.Props,
		Text:           t.Text,
		HeaderControls: NewHeaderControlProps(t.Width, t.Height),
	}
}

// Draw renders the terminal window.
func (t *Terminal) Draw() string {
	result, err := RenderTemplate("terminal", t.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering terminal template: %v -->", err)
	}
	return result
}
