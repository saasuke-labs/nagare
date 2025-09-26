# Nagare

## Examples

### Browser and VM

This is still in progress

```text
browser:Browser@home
vps:VM@ubuntu {
    nginx:App
    app:App
}

@home(url: "https://www.nagare.com", bg: "#e6f3ff", fg: "#333", text: "Home Page")
@ubuntu(title: "home@ubuntu", bg: "darkorange", fg: "#333", text: "Ubuntu")
```

Becomes:

![Browser and VM](static/examples/example2.svg)

(grid is there to help me why developing)
