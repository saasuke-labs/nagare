package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/saasuke-labs/nagare/pkg/chart"
	"github.com/saasuke-labs/nagare/pkg/diagram"
)

func main() {
	http.HandleFunc("POST /render", handleRender)
	http.HandleFunc("POST /render-webp", handleRenderWebP)
	http.HandleFunc("GET /test", handleTest)
	log.Printf("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
	html, err := diagram.CreateDiagram(code)
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

	input := strings.TrimSpace(string(code))
	var html string

	// Check first line to determine type
	if strings.HasPrefix(input, "chart") {
		// Parse and render chart
		c, err := chart.Parse(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		html = c.RenderHTML()
	} else {
		// Default to diagram
		html, err = diagram.CreateDiagram(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Send response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleRenderWebP(w http.ResponseWriter, r *http.Request) {
	code, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	data, err := diagram.CreateDiagramWebP(string(code))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Write(data)
}
