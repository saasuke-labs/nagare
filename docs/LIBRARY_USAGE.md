# Nagare Library Usage Examples

This document provides examples of using nagare as a library in your Go projects.

## Installation

```bash
go get github.com/saasuke-labs/nagare
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/saasuke-labs/nagare/pkg/nagare"
)

func main() {
    diagramCode := `
@layout(w:500,h:300)

client:Browser@webapp
server:Server@backend

client.e --> server.w

@client(x:50,y:100,w:180,h:120)
@webapp(url: "https://example.com", bg: "#e6f3ff", fg: "#333", text: "Web App")

@server(x:300,y:100,w:150,h:50, title: "API Server", icon: "server", port: 8080, bg: "#f0f8ff", fg: "#333")
`

    // Render to SVG string
    svg, err := nagare.RenderToSVG(diagramCode)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Println(svg)
}
```

## Render File to File

```go
package main

import (
    "log"

    "github.com/saasuke-labs/nagare/pkg/nagare"
)

func main() {
    err := nagare.RenderFileToFile("input.nagare", "output.svg")
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
}
```

## With Debug Output

```go
package main

import (
    "log"

    "github.com/saasuke-labs/nagare/pkg/nagare"
)

func main() {
    diagramCode := `
@layout(w:500,h:300)
client:Browser@webapp
server:Server@backend
client.e --> server.w
`

    // This will print debug info to stdout
    svg, err := nagare.RenderToSVGWithDebug(diagramCode)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    // svg now contains the rendered diagram
    _ = svg
}
```

## HTTP Server Example

```go
package main

import (
    "io"
    "log"
    "net/http"

    "github.com/saasuke-labs/nagare/pkg/nagare"
)

func handleRender(w http.ResponseWriter, r *http.Request) {
    code, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
    }

    svg, err := nagare.RenderToSVG(string(code))
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "image/svg+xml")
    w.Write([]byte(svg))
}

func main() {
    http.HandleFunc("POST /render", handleRender)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## API Reference

### `RenderToSVG(code string) (string, error)`

Takes nagare DSL code as a string and returns the rendered SVG as a string. This is the main entry point for using nagare as a library.

**Parameters:**

- `code`: The nagare DSL code to render

**Returns:**

- `string`: The rendered SVG
- `error`: Any error that occurred during rendering

### `RenderToSVGWithDebug(code string) (string, error)`

Like `RenderToSVG` but prints debug information to stdout. Useful during development.

### `RenderFileToFile(inputPath, outputPath string) error`

Reads a nagare file from `inputPath` and writes the SVG output to `outputPath`.

**Parameters:**

- `inputPath`: Path to the input .nagare file
- `outputPath`: Path where the output .svg file will be written

**Returns:**

- `error`: Any error that occurred during rendering or file I/O
