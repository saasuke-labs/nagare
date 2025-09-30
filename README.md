# Nagare

## Examples

Pull requests automatically run a preview workflow that boots the Nagare server, renders the `/test` diagram, and attaches the resulting SVG to the PR discussion for quick review.

## Layout Overrides

You can control the overall canvas dimensions with a global `@layout` directive. This is useful when you need extra room for connections or when you want diagrams to render inside a specific viewport.

```text
@layout(w: 800, h: 600)

browser:Browser@home
vps:VM@ubuntu {
    nginx:App
    app:App
}
```

The `layout` stage resolves these geometry overrides before components are instantiated, so every downstream step (child placement, connection routing, SVG rendering) respects the requested dimensions.

### Browser and VM

This is still in progress

```text
browser:Browser@home
vps:VM@ubuntu {
    nginx:App
    app:App
}

@home(url: "https://www.nagare.com", bg: "#e6f3ff", fg: "#333", text: "Home Page")
@ubuntu(title: "home@ubuntu", bg: "darkorange", fg: "#333", contentBg: "#ccc")
```

Becomes:

![Browser and VM](static/examples/example2.svg)

(grid is there to help me why developing)

## Development

The Go module under `nagare/` powers the HTTP server and static rendering pipeline. Format and test Go code before sending a pull request:

```bash
cd nagare
go test ./...
```

The layout unit tests describe how connection routing, geometry inheritance, and canvas bounds interact. They are a good starting point for understanding how new components should integrate with the existing pipeline.
