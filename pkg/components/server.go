package components

import (
	"bytes"
	"fmt"
	"text/template"

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
func (s *Server) Draw() string {
	actualWidth := s.Width
	actualHeight := s.Height

	data := struct {
		X      float64
		Y      float64
		Width  float64
		Height float64
		Props  ServerProps
		Text   string
	}{
		X:      s.X,
		Y:      s.Y,
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
        <path d="M-33.442 42.023v-7.637a.68.68 0 0 1 .385-.651l13.173-7.608c.237-.148.503-.178.74-.03l13.232 7.637a.71.71 0 0 1 .355.651V49.63a.71.71 0 0 1-.355.651l-11.367 6.57a56.27 56.27 0 0 1-1.806 1.036c-.266.148-.533.148-.8 0l-13.202-7.608c-.237-.148-.355-.326-.355-.622v-7.637z" fill="#009438"/><path d="M-24.118 39.18v8.9c0 1.006-.8 1.894-1.865 1.865-.65-.03-1.154-.296-1.5-.858-.178-.266-.237-.562-.237-.888V35.836c0-.83.503-1.42 1.154-1.687s1.302-.207 1.954 0c.622.178 1.095.562 1.5 1.036l7.874 9.443c.03.03.06.09.118.148v-9c0-.947.65-1.687 1.57-1.776 1.154-.148 1.924.68 2.042 1.54v12.6c0 .7-.326 1.214-.918 1.54-.444.237-.918.296-1.42.266a3.23 3.23 0 0 1-1.954-.829c-.296-.266-.503-.592-.77-.888l-7.49-8.97c0-.03-.03-.06-.06-.09z" fill="#fefefe"/>
        {{else if eq .Props.Icon "golang"}}
        <!-- Golang icon -->
        <style>.st0{fill:#00acd7}</style><switch><g><path class="st0" d="M22.3 24.7c-.1 0-.2-.1-.1-.2l.7-1c.1-.1.2-.2.4-.2h12.6c.1 0 .2.1.1.2l-.6.9c-.1.1-.2.2-.4.2l-12.7.1zM17 27.9c-.1 0-.2-.1-.1-.2l.7-1c.1-.1.2-.2.4-.2h16.1c.1 0 .2.1.2.2l-.3 1c0 .1-.2.2-.3.2H17zm8.5 3.3c-.1 0-.2-.1-.1-.2l.5-.9c.1-.1.2-.2.4-.2h7c.1 0 .2.1.2.2l-.1.8c0 .1-.1.2-.2.2l-7.7.1zM62.1 24l-5.9 1.5c-.5.1-.6.2-1-.4-.5-.6-.9-1-1.7-1.3-2.2-1.1-4.4-.8-6.4.5-2.4 1.5-3.6 3.8-3.6 6.7 0 2.8 2 5.1 4.8 5.5 2.4.3 4.4-.5 6-2.3.3-.4.6-.8 1-1.3h-6.8c-.7 0-.9-.5-.7-1.1.5-1.1 1.3-2.9 1.8-3.8.1-.2.4-.6.9-.6h12.8c-.1 1-.1 1.9-.2 2.9-.4 2.5-1.3 4.9-2.9 6.9-2.5 3.3-5.8 5.4-10 6-3.5.5-6.7-.2-9.5-2.3-2.6-2-4.1-4.6-4.5-7.8-.5-3.8.7-7.3 3-10.3 2.5-3.3 5.8-5.4 9.9-6.1 3.3-.6 6.5-.2 9.3 1.7 1.9 1.2 3.2 2.9 4.1 5 .1.4 0 .5-.4.6z"/><path class="st0" d="M73.7 43.5c-3.2-.1-6.1-1-8.6-3.1-2.1-1.8-3.4-4.1-3.8-6.8-.6-4 .5-7.5 2.9-10.6 2.6-3.4 5.7-5.1 9.9-5.9 3.6-.6 7-.3 10 1.8 2.8 1.9 4.5 4.5 5 7.9.6 4.8-.8 8.6-4 11.9-2.3 2.4-5.2 3.8-8.4 4.5-1.1.2-2.1.2-3 .3zm8.4-14.2c0-.5 0-.8-.1-1.2-.6-3.5-3.8-5.5-7.2-4.7-3.3.7-5.4 2.8-6.2 6.1-.6 2.7.7 5.5 3.2 6.7 1.9.8 3.9.7 5.7-.2 2.9-1.4 4.4-3.7 4.6-6.7z"/></g></switch>
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
