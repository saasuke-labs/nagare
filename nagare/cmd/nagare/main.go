package main

import (
	"fmt"
	"io"
	"log"
	"nagare/layout"
	"nagare/parser"
	"nagare/renderer"
	"nagare/tokenizer"
	"net/http"
)

func main() {
	http.HandleFunc("POST /render", handleRender)
	http.HandleFunc("GET /test", handleTest)
	log.Printf("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createDiagram(code string) (string, error) {
	fmt.Printf("Input code:\n%s\n", string(code))

	// Pipeline:
	// 1. Tokenize
	tokens := tokenizer.Tokenize(string(code))
	fmt.Printf("Tokens: %+v\n", tokens)

	// 2. Parse
	ast, err := parser.Parse(tokens)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	fmt.Printf("AST: \n%+v\n", ast)

	// 3. Layout
	const defaultCanvasWidth, defaultCanvasHeight = 800.0, 400.0
	l := layout.Calculate(ast, defaultCanvasWidth, defaultCanvasHeight)

	fmt.Printf("Layout: \n%+v\n", l)

	// 4. Render using the computed layout dimensions
	canvasWidth := int(l.Bounds.Width)
	canvasHeight := int(l.Bounds.Height)
	if canvasWidth == 0 {
		canvasWidth = int(defaultCanvasWidth)
	}
	if canvasHeight == 0 {
		canvasHeight = int(defaultCanvasHeight)
	}

	html := renderer.Render(l, canvasWidth, canvasHeight)
	fmt.Println(html)
	return html, nil
}

func handleTest(w http.ResponseWriter, r *http.Request) {

	code := `
@layout(w:950,h:400)

	
browser:Browser@home
vps:VM@ubuntu {
    nginx:Server@nginx
    app:Server@app
}

@browser(x:50,y:175,w:200,h:150)
@vps(x:300,y:50,w:600,h:300)

@home(url: "https://www.nagare.com", bg: "#e6f3ff", fg: "#333", text: "Home Page")
@ubuntu(title: "home@ubuntu", bg: "darkorange", fg: "#333", text: "Ubuntu")
@nginx(x:50,y:125,w:200,h:50, title: "nginx", icon: "nginx", port: 80, bg: "#e6f3ff", fg: "#333")
@app(x:350,y:125,w:200,h:50, title: "App", icon: "golang", port: 8080, bg: "#f0f8ff", fg: "#333")
`
	html, err := createDiagram(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	// Read the input
	code, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	html, err := createDiagram(string(code))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
