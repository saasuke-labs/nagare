# Agent Notes for `nagare/`

## Scope

- Work within the main Go module in the root directory.
- Focus on maintaining a clean, well-structured codebase following Go best practices.

## Architecture Overview

Nagare focuses on **text-to-SVG** rendering in this repository. WebP/image export is intentionally out of scope and handled externally.

The project follows a pipeline architecture with four main stages:

1. **Tokenization** (`pkg/tokenizer/`)

   - Converts DSL text into typed tokens (identifiers, braces, @-states, properties, etc.)

2. **Parsing** (`pkg/parser/`)

   - Builds an AST from tokens
   - Handles `@state(...)` blocks and stores properties
   - Currently supports single-level nesting

3. **Layout** (`pkg/layout/`)

   - Processes the AST and instantiates components
   - Handles geometry calculations and grid-based positioning
   - Applies state properties to components
   - Manages component relationships and connections

4. **Rendering** (`pkg/renderer/`)
   - Components render themselves to SVG
   - Combines individual SVG fragments into final output

### Key Components

- `cmd/main.go` - HTTP server with `/render` (POST) and `/test` (GET) endpoints
- `pkg/components/` - SVG component definitions (Browser, VM, Server, etc.)
- `pkg/props/` - Property parsing helpers for component configuration
- `pkg/version/` - Version information and build metadata

## Coding Conventions

- Go code must be formatted with `gofmt` before committing.
- Organize new logic by extending the existing pipeline stages (tokenizer → parser → layout → renderer) rather than skipping around them.
- Reuse prop parsing helpers in `props/`; avoid duplicating parsing logic inside components.
- In the `layout/` package, leverage helpers like `applyIDStateProperties`, `applyNamedStateProperties`, and `routeArrowPoints` when adding new component types or connection rules so geometry/state handling remains consistent.

- When adding or refactoring diagram components under `pkg/diagram/components/`, keep component-specific render-node conversion in that package (for example, `DrawFromRenderNode`) so `pkg/diagram/diagram.go` stays orchestration-only.
- Render-node conversion must treat component drawing coordinates as local to the render-tree group transform (set `x/y` to local origin and let recursive `<g transform>` handle placement) to avoid doubled offsets.

## Development Workflow

```bash
# Run tests
go test ./...

# Start development server
go run ./cmd/main.go

# Build release binary
goreleaser build --snapshot --clean

# Add CI/pr diagrams
# Create `.nagare` files under `.github/testdiagrams/`
```

### CI Rendering Notes

- Keep Nagare runtime output focused on SVG.
- CI preview artifacts should stay SVG-only; do not add WebP conversion steps or WebP rendering logic back into this repository.

## Known Limitations

- Component types in `layout/Calculate` require manual registration
- Single level of nesting in containers
- State property parsing could be more flexible

## Future Improvements

- Support for deeper nesting levels
- Dynamic component type registration
- More flexible property parsing
- Additional component types
- Enhanced connection routing
