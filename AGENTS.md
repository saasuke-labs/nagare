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
- `static/playground.html` - Browser playground UI (editor, live render, localStorage-backed save/load + autosave behavior)
- `pkg/diagram/components/` - Component-specific render-node translation and rendering entrypoints
- `pkg/components/` - Shared SVG drawing primitives/templates used by diagram component adapters
- `pkg/props/` - Property parsing helpers for component configuration
- `pkg/version/` - Version information and build metadata

## Coding Conventions

- Go code must be formatted with `gofmt` before committing.
- Organize new logic by extending the existing pipeline stages (tokenizer → parser → layout → renderer) rather than skipping around them.
- Reuse prop parsing helpers in `props/`; avoid duplicating parsing logic inside components.
- Prefer composing higher-level diagram components from primitive render components (for example `Database` composed from `Cylinder` + `Led`) when behavior can be shared across types.
- When an action is semantically an alias, map it at the parent component boundary (for example `Database.read/write` -> child `Led.blink`) so primitives keep a small, stable action surface.
- For action-generated arrows, map action semantics at the source component boundary (for example `Browser.request/response` deciding solid vs dashed style) and instantiate Arrow with resolved values; Arrow itself should remain action-agnostic.
- Parser should remain action-name agnostic: do not hardcode component action names in `pkg/parser`; action-to-arrow mapping belongs in component boundaries and downstream orchestration/layout.
- Prefer inline component attributes (`name:Type(x:..., y:...)`) for concise examples, while keeping `@id(...)` and `name:Type@state` as equally supported syntax choices; keep `@layout(...)` and action states (for example `@db.read(...)`) for global layout and action timelines.
- Anonymous inline declarations like `Type(x:...)` are valid and receive an auto-generated id in the parser.
- For composed visuals, prefer parent-relative/proportional child geometry (instead of fixed pixel offsets) so composition scales correctly when parent dimensions change.
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

### Deployment CI Notes

- Cloud Run deployment CI uses GitHub OIDC + Google Workload Identity Federation.
- If deployment auth fails with `iam.serviceAccounts.getAccessToken`, verify the deployer service account grants both `roles/iam.workloadIdentityUser` and `roles/iam.serviceAccountTokenCreator` to the GitHub `principalSet`.
- Keep deployment setup docs in sync in `docs/GCP_CLOUD_RUN_CI_SETUP.md` whenever workflow auth/deploy settings change.

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


### RFC references

- Action-driven arrow creation and centralized arrow resolution plan: `docs/RFC_PORT_AWARE_ARROWS.md`
