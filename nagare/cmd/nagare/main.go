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
	http.HandleFunc("/render", handleRender)
	log.Printf("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request")
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

	fmt.Printf("Input code:\n%s\n", string(code))

	// Pipeline:
	// 1. Tokenize
	tokens := tokenizer.Tokenize(string(code))
	fmt.Printf("Tokens: %+v\n", tokens)

	// 2. Parse
	ast, err := parser.Parse(tokens)
	if err != nil {
		http.Error(w, "Parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("AST: \n%+v\n", ast)

	// 3. Layout
	const canvasWidth, canvasHeight = 800, 400
	l := layout.Calculate(ast, canvasWidth, canvasHeight)

	fmt.Printf("Layout: \n%+v\n", l)

	// 4. Render
	html := renderer.Render(l, canvasWidth, canvasHeight)
	fmt.Println(html)
	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
