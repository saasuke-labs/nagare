# Justfile for nagare

# Default recipe - show available commands
default:
    @just --list

# Start the server on port 8080
start port="8080":
    go run ./cmd/main.go -port {{port}}

# Start the server with hot reload (requires air: go install github.com/air-verse/air@latest)
start-watch port="8080":
    $(go env GOPATH)/bin/air

# Build the binary
build:
    go build -o nagare ./cmd/main.go

# Generate SVG from input file
gen input output="output.svg":
    go run ./cmd/main.go -input {{input}} -output {{output}}

# Run tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Format code
fmt:
    gofmt -w .

# Install development tools
install-tools:
    @echo "Installing air for hot reload..."
    go install github.com/air-verse/air@latest
    @echo "Installing goreleaser..."
    go install github.com/goreleaser/goreleaser@latest
    @echo ""
    @echo "✅ Done! Tools installed to $(go env GOPATH)/bin"
    @echo ""
    @echo "💡 Note: If you want to use 'air' command directly, add this to your ~/.zshrc:"
    @echo "   export PATH=\"\$$PATH:\$$(go env GOPATH)/bin\""

# Clean build artifacts
clean:
    rm -f nagare
    rm -f coverage.out
    rm -rf dist/
    rm -rf tmp/
