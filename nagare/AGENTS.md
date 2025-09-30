# Agent Notes for `nagare/`

## Scope
- Only modify files inside this Go module (`/workspace/nagare/nagare`).
- Treat sibling directories at the repo root (e.g., `server/`, `nagare-go/`, NodeJS prototypes) as out of scope unless explicitly asked.

## Architecture Overview
- `cmd/nagare/main.go` exposes `/render` (POST) and `/test` (GET) endpoints. Both run the same four-stage pipeline and respond with generated HTML/SVG.
- The pipeline is strictly ordered: `tokenizer` ➜ `parser` ➜ `layout` ➜ `renderer`.
  - `tokenizer/` turns the DSL text into typed tokens (identifiers, braces, @-states, prop punctuation, etc.).
  - `parser/` builds a shallow AST (max depth 1) with optional named states defined via `@state(...)` blocks whose props are stored for reuse.
  - `layout/` walks the AST, instantiates concrete component types (e.g., Browser, VM) and assigns grid-based geometry (48-column grid, coordinates in component `Shape`). Props from referenced states are parsed and attached to each component instance.
  - `renderer/` asks components to render themselves, producing SVG fragments combined into the final document.
- `components/` contains the SVG/component definitions. Each component carries a `Shape`, optional state metadata, and a `Render` method that returns SVG strings.
- `props/` provides helpers that parse `key:value` pairs (with quoted strings, ints, etc.) into strongly-typed prop structs used by components.

## Coding Conventions
- Go code must be formatted with `gofmt` before committing.
- Organize new logic by extending the existing pipeline stages (tokenizer → parser → layout → renderer) rather than skipping around them.
- Reuse prop parsing helpers in `props/`; avoid duplicating parsing logic inside components.
- In the `layout/` package, leverage helpers like `applyIDStateProperties`, `applyNamedStateProperties`, and `routeArrowPoints` when adding new component types or connection rules so geometry/state handling remains consistent.

## Useful Commands
- `go test ./...` — run the unit test suite (currently sparse, but keeps the module compiling).
- `go run ./cmd/nagare` — starts the HTTP server on `localhost:8080` and exercises the sample DSL via `/test`.

## Known Limitations
- Component type handling in `layout/Calculate` is currently hard-coded for a few types (Browser, VM, Server, etc.); adding a type requires updating this switch and providing a component implementation.
- Only one level of nesting is supported in the parser; containers cannot contain nested containers beyond depth 1.
