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
	const canvasWidth, canvasHeight = 800, 400
	l := layout.Calculate(ast, canvasWidth, canvasHeight)

	fmt.Printf("Layout: \n%+v\n", l)

	// 4. Render
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

@browser(x:50,y:50,w:300,h:300)
@vps(x:400,y:50,w:550,h:300)
@nginx(x:475,y:150,w:200,h:150)
@app(x:700,y:150,w:200,h:150)

@home(url: "https://www.nagare.com", bg: "#e6f3ff", fg: "#333", text: "Home Page")
@ubuntu(title: "home@ubuntu", bg: "darkorange", fg: "#333", text: "Ubuntu")
@nginx(title: "Nginx Server", icon: "nginx", port: 80, bg: "#e6f3ff", fg: "#333")
@app(title: "App Server", icon: "golang", port: 8080, bg: "#f0f8ff", fg: "#333")
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
