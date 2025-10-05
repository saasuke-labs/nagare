package components

import (
	"fmt"

	"github.com/saasuke-labs/nagare/pkg/props"
)

// ServerProps defines the configurable properties for a Server component
type ServerProps struct {
	Title           string `prop:"title"`
	Icon            string `prop:"icon"` // nginx, golang, etc.
	Port            int    `prop:"port"`
	BackgroundColor string `prop:"bg"`
	ForegroundColor string `prop:"fg"`
}

// Parse implements the Props interface
func (s *ServerProps) Parse(input string) error {
	return props.ParseProps(input, s)
}

// DefaultServerProps returns a ServerProps with default values
func DefaultServerProps() ServerProps {
	return ServerProps{
		Title:           "",
		Icon:            "default",
		Port:            80,
		BackgroundColor: "#e6f3ff",
		ForegroundColor: "#333333",
	}
}

type Server struct {
	Rectangle
	Props ServerProps
	State string // Current state name
}

type ServerTemplateData struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Props  ServerProps
	Text   string
}

// NewServer creates a new Server with default props
func NewServer(id string) *Server {
	return &Server{
		Rectangle: Rectangle{
			Text: id,
		},
		Props: DefaultServerProps(),
	}
}

func (s *Server) templateData() ServerTemplateData {
	return ServerTemplateData{
		X:      s.X,
		Y:      s.Y,
		Width:  s.Width,
		Height: s.Height,
		Props:  s.Props,
		Text:   s.Text,
	}
}

// Render returns an SVG representation of the server
func (s *Server) Render() (string, error) {
	result, err := RenderTemplate("server", s.templateData())
	if err != nil {
		return "", fmt.Errorf("error rendering template: %w", err)
	}

	return result, nil
}

// Configure sets the server's properties from a string
func (s *Server) Configure(props string) error {
	return s.Props.Parse(props)
}

// Draw implements the Component interface
func (s *Server) Draw() string {
	result, err := RenderTemplate("server", s.templateData())
	if err != nil {
		return fmt.Sprintf("<!-- Error rendering template: %v -->", err)
	}

	return result
}
