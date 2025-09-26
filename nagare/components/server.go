package components

import (
	"bytes"
	"fmt"
	"nagare/props"
	"text/template"
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

// NewServer creates a new Server with default props
func NewServer(id string) *Server {
	return &Server{
		Rectangle: Rectangle{
			Text: id,
		},
		Props: DefaultServerProps(),
	}
}

// Render returns an SVG representation of the server
func (s *Server) Render() (string, error) {
	tmpl, err := template.New("server").Parse(ServerTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}

	var svg bytes.Buffer
	err = tmpl.Execute(&svg, s)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return svg.String(), nil
}

// Configure sets the server's properties from a string
func (s *Server) Configure(props string) error {
	return s.Props.Parse(props)
}

// Draw implements the Component interface
func (s *Server) Draw(colWidth, rowHeight float64) string {
	actualWidth := float64(s.Width) * colWidth
	actualHeight := float64(s.Height) * rowHeight

	data := struct {
		X      float64
		Y      float64
		Width  float64
		Height float64
		Props  ServerProps
		Text   string
	}{
		X:      float64(s.X) * colWidth,
		Y:      float64(s.Y) * rowHeight,
		Width:  actualWidth,
		Height: actualHeight,
		Props:  s.Props,
		Text:   s.Text,
	}

	// Create template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b float64) float64 { return a + b },
		"mul": func(a, b float64) float64 { return a * b },
		"sub": func(a, b float64) float64 { return a - b },
	}

	tmpl, err := template.New("server").Funcs(funcMap).Parse(ServerTemplate)
	if err != nil {
		return fmt.Sprintf("<!-- Error parsing template: %v -->", err)
	}

	var svg bytes.Buffer
	err = tmpl.Execute(&svg, data)
	if err != nil {
		return fmt.Sprintf("<!-- Error executing template: %v -->", err)
	}

	return svg.String()
}

const ServerTemplate = `
<g transform="translate({{printf "%.6f" .X}},{{printf "%.6f" .Y}})">
    <!-- Main server box -->
    <rect x="0" y="0" width="{{printf "%.6f" .Width}}" height="{{printf "%.6f" .Height}}" 
          rx="{{printf "%.6f" (mul .Height 0.1)}}" ry="{{printf "%.6f" (mul .Height 0.1)}}" 
          fill="{{.Props.BackgroundColor}}" stroke="{{.Props.ForegroundColor}}" stroke-width="2"/>
    
    <!-- Icon -->
    {{$iconSize := mul .Height 0.7}}
    {{$iconMargin := mul .Height 0.15}}
    <g transform="translate({{printf "%.6f" $iconMargin}},{{printf "%.6f" $iconMargin}})">
        {{if eq .Props.Icon "nginx"}}
        <!-- Nginx icon -->
        <rect x="0" y="0" width="{{printf "%.6f" $iconSize}}" height="{{printf "%.6f" $iconSize}}" fill="#009639"/>
        <text x="{{printf "%.6f" (mul $iconSize 0.5)}}" y="{{printf "%.6f" (mul $iconSize 0.7)}}" 
              fill="#ffffff" font-family="Arial" font-size="{{printf "%.6f" (mul $iconSize 0.6)}}" 
              text-anchor="middle">N</text>
        {{else if eq .Props.Icon "golang"}}
        <!-- Golang icon -->
        <rect x="0" y="0" width="{{printf "%.6f" $iconSize}}" height="{{printf "%.6f" $iconSize}}" fill="#00ADD8"/>
        <text x="{{printf "%.6f" (mul $iconSize 0.5)}}" y="{{printf "%.6f" (mul $iconSize 0.7)}}" 
              fill="#ffffff" font-family="Arial" font-size="{{printf "%.6f" (mul $iconSize 0.6)}}" 
              text-anchor="middle">Go</text>
        {{else}}
        <!-- Default server icon -->
        <rect x="0" y="0" width="{{printf "%.6f" $iconSize}}" height="{{printf "%.6f" $iconSize}}" fill="#666666"/>
        <text x="{{printf "%.6f" (mul $iconSize 0.5)}}" y="{{printf "%.6f" (mul $iconSize 0.7)}}" 
              fill="#ffffff" font-family="Arial" font-size="{{printf "%.6f" (mul $iconSize 0.6)}}" 
              text-anchor="middle">S</text>
        {{end}}
    </g>
    
    <!-- Title -->
    <text x="{{printf "%.6f" (add (mul .Height 1.0) $iconMargin)}}" 
          y="{{printf "%.6f" (mul .Height 0.6)}}" 
          fill="{{.Props.ForegroundColor}}" 
          font-family="Arial" 
          font-size="{{printf "%.6f" (mul .Height 0.4)}}"
          >{{.Props.Title}}</text>
    
    <!-- Port display -->
    <text x="{{printf "%.6f" (sub .Width (mul .Height 0.15))}}" 
          y="{{printf "%.6f" (mul .Height 0.6)}}" 
          fill="{{.Props.ForegroundColor}}" 
          font-family="Arial" 
          font-size="{{printf "%.6f" (mul .Height 0.35)}}"
          text-anchor="end">:{{.Props.Port}}</text>
</g>`
