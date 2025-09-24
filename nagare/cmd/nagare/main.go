package main

import (
	"io"
	"log"
	"nagare/layout"
	"nagare/parser"
	"nagare/renderer"
	"nagare/tokenizer"
	"net/http"
)

func main() {
	http.HandleFunc("/render", handleRender)
	log.Printf("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the input
	code, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Pipeline:
	// 1. Tokenize
	tokens := tokenizer.Tokenize(string(code))

	// 2. Parse
	ast := parser.Parse(tokens)

	// 3. Layout
	const canvasWidth, canvasHeight = 400, 300
	l := layout.Calculate(ast, canvasWidth, canvasHeight)

	// 4. Render
	html := renderer.Render(l, canvasWidth, canvasHeight)

	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
