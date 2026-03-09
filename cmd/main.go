package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/saasuke-labs/nagare/pkg/nagare"
)

func main() {
	// Define CLI flags
	inputFile := flag.String("input", "", "Input .nagare file path")
	outputFile := flag.String("output", "", "Output file path (e.g., diagram.svg)")
	format := flag.String("format", "", "Output format: svg (auto-detected from output file extension if not specified)")
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	// CLI mode: if input and output are provided, render file and exit
	if *inputFile != "" && *outputFile != "" {
		if err := renderFile(*inputFile, *outputFile, *format); err != nil {
			log.Fatalf("Error rendering diagram: %v", err)
		}
		fmt.Printf("Successfully rendered %s to %s\n", *inputFile, *outputFile)
		return
	}

	// HTTP server mode: if no CLI args provided, start the server
	if *inputFile == "" && *outputFile == "" {
		// Playground UI
		http.HandleFunc("GET /", handlePlayground)
		http.HandleFunc("POST /api/render", handleAPIRender)

		// Legacy endpoints (backward compatibility)
		http.HandleFunc("POST /render", handleRender)
		http.HandleFunc("GET /test", handleTest)

		log.Printf("Server starting on http://localhost:%s", *port)
		log.Printf("Playground UI available at http://localhost:%s/", *port)
		log.Fatal(http.ListenAndServe(":"+*port, nil))
		return
	}

	// Invalid usage
	fmt.Fprintln(os.Stderr, "Error: Both -input and -output flags must be provided together for CLI mode")
	fmt.Fprintln(os.Stderr, "\nUsage:")
	fmt.Fprintln(os.Stderr, "  CLI mode:    nagare -input diagram.nagare -output diagram.svg [-format svg]")
	fmt.Fprintln(os.Stderr, "  Server mode: nagare [-port 8080]")
	os.Exit(1)
}

func handleTest(w http.ResponseWriter, r *http.Request) {

	code := `
@layout(w:950,h:400)

browser:Browser@home
vps:VM@ubuntu {
nginx:Server@nginx
app:Server@app
}

browser.e --> nginx.w
nginx.e --> app.w

@browser(x:50,y:100,w:200,h:150)
@home(url: "https://www.nagare.com", bg: "#e6f3ff", fg: "#333", text: "Home Page")

@vps(x:300,y:&browser.c,w:600,h:300)
@ubuntu(title: "home@ubuntu", bg: "#333", fg: "#ccc", text: "Ubuntu")

@nginx(x:50,y:&browser.c,w:200,h:50, title: "nginx", icon: "nginx", port: 80, bg: "#e6f3ff", fg: "#333")
@app(x:350,y:&browser.c,w:200,h:50, title: "App", icon: "golang", port: 8080, bg: "#f0f8ff", fg: "#333")
`
	html, err := nagare.RenderToSVG(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handlePlayground(w http.ResponseWriter, r *http.Request) {
	// Serve the playground HTML file
	http.ServeFile(w, r, "static/playground.html")
}

func handleAPIRender(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	// Render to SVG
	svg, err := nagare.RenderToSVG(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	elapsed := time.Since(start)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("X-Render-Time", strconv.FormatInt(elapsed.Milliseconds(), 10))
	w.Write([]byte(svg))
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	// Read the input
	code, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	input := strings.TrimSpace(string(code))
	html, err := nagare.RenderToHTML(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// renderFile renders an input file to an output file in the specified format.
// Format can only be "svg". If format is empty, it's auto-detected from the output file extension.
func renderFile(inputPath, outputPath, format string) error {
	// Auto-detect format from file extension if not specified
	if format == "" {
		format = "svg"
	}

	// Read input file
	code, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	switch strings.ToLower(format) {
	case "svg":
		return renderToSVGFile(string(code), outputPath)
	default:
		return fmt.Errorf("unsupported format: %s (use 'svg')", format)
	}
}

// renderToSVGFile renders code to an SVG file
func renderToSVGFile(code, outputPath string) error {
	svg, err := nagare.RenderToSVG(code)
	if err != nil {
		return fmt.Errorf("failed to render diagram: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(svg), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
