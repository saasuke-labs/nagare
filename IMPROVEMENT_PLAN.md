# Nagare Code Improvement Plan

## Background

This document is a phased refactoring plan for the Nagare codebase. Each phase is
designed to be handed off to an independent agent. Phases build on each other but
each leaves the codebase in a fully working, passing-test state. Run `go test ./...`
after every phase to verify nothing is broken.

### Current pipeline (for orientation)

```
DSL text
  → pkg/tokenizer   Tokenize()  → []Token
  → pkg/parser      Parse()     → parser.Node (AST + Connections + States)
  → pkg/layout      Calculate() → layout.Layout (NodeIndex + Children + Arrows)
  → pkg/diagram     BuildRenderTree() → *RenderNode tree
  → pkg/diagram/components/*  DrawFromRenderNode() → SVG strings
  → renderer        Render()    → final <svg>…</svg>
Public entry-point: pkg/nagare.RenderToSVG()
HTTP / CLI shell:   cmd/main.go
```

### High-level catalogue of issues

| # | Category | Where |
|---|----------|-------|
| 1 | Debug `fmt.Print*` in production paths | props.go, browser.go, infrastructure.go, layout.go (many), diagram.go |
| 2 | "Legacy" as a production type name | every `pkg/diagram/components/*/xxx.go`, `core/types.go` |
| 3 | Bloated central orchestrator | `pkg/layout/layout.go` (1 342 lines, 4 type-switches, 15 nearly-identical build functions) |
| 4 | Fragmented / duplicated geometry types | `components.Shape` vs `core.BoundingBox`; `layout.Point` vs `core.Point` |
| 5 | `map[string]any` as cross-stage data carrier | RenderNode.Props, DiagramNode.Props, DrawFromRenderNode args |
| 6 | Duplicate utility functions | `actionMapFromAny`, `floatProp`/`FloatProp`, `splitProps`/`ParseProps` |
| 7 | Broken API boundary | `cmd/main.go` imports `pkg/diagram` directly; public API `pkg/nagare` is bypassed |
| 8 | Silent error swallowing | `_ = comp.Props.Parse(raw)`, `fmt.Printf("failed to …")` everywhere |

---

## Phase 1 — Eliminate debug output from production code

### Goal
Remove every `fmt.Print*` call that is not a deliberate, user-facing log line. After
this phase the public API produces no standard-output side-effects.

### Motivation
Debug prints make the library useless as a dependency (callers cannot suppress
them), pollute server logs, and mislead anyone reading the code about what is
actually computed versus printed.

### Files to change

#### `pkg/props/props.go`
- Remove lines:
  ```go
  fmt.Printf("Parsing prop key=%q value=%q\n", key, value)
  fmt.Printf("After cleanup: key=%q value=%q\n", key, value)
  fmt.Printf("Setting field %s with tag %q to %q\n", field.Name, prop, value)
  fmt.Printf("Warning: no field found with prop tag %q\n", key)
  ```
- Remove the `"fmt"` import if it is no longer used.

#### `pkg/components/browser.go`
- Remove:
  ```go
  fmt.Println("Drawing browser at", r.X, r.Y, "size", r.Width, r.Height)
  ```
- Remove `fmt.Printf("Error rendering template: %v\n", err)` — replace it with
  `return ""` or, better, return the error message as an SVG comment
  (already done in some components as
  `return fmt.Sprintf("<!-- Error rendering … -->")`).

#### `pkg/components/infrastructure.go`  (Database.Draw)
- Remove:
  ```go
  fmt.Println("------------------------")
  fmt.Println("Rendering database: ", d, d.templateData())
  fmt.Println("Result:", result)
  ```
- The `err` branch already returns a comment string; keep that pattern.

#### `pkg/layout/layout.go`
Remove or replace every `fmt.Print*` found in this file. The complete list is:
- `fmt.Printf("Alignment reference detected for X: …")` (line ~153)
- `fmt.Printf("Possible broken alignment reference detected for X: …")` (line ~160)
- `fmt.Printf("Alignment reference detected for Y: …")` (line ~176)
- `fmt.Printf("Possible broken alignment reference detected for Y: …")` (line ~183)
- `fmt.Printf("Failed to resolve alignment reference …")` (line ~272)
- `fmt.Printf("Resolved X alignment for %s: …")` (line ~279)
- `fmt.Printf("Resolved Y alignment for %s: …")` (line ~283)
- `fmt.Printf("Failed to resolve reconstructed alignment reference …")` (line ~296)
- `fmt.Printf("Resolved Y alignment (reconstructed) …")` (line ~301)
- `fmt.Printf("State: %s, Props: %+v\n", …)` in buildBrowser, buildVM (lines ~600, ~618)
- `fmt.Printf("failed to parse @layout props: …")` (line ~538)
- `fmt.Printf("failed to parse geometry for %s: …")` (line ~1008)
- `fmt.Printf("failed to parse props for %s: …")` (line ~1020)
- `fmt.Printf("connection skipped: …")` (line ~1049)
- `fmt.Printf("Unknown child type: %s\n", …)` in layoutVMChildren (line ~671)

For now, replace each with `// silently ignore` or a no-op; proper error
propagation is addressed in Phase 7.

#### `pkg/diagram/diagram.go`
- Remove `fmt.Println(svg)` from `CreateDiagram()` (line ~592). This is a
  significant bug: every call to the public API prints the full SVG to stdout.

#### Imports cleanup
After removing all `fmt.Print*` calls, run `go build ./...`; remove any `"fmt"`
import that the compiler flags as unused.

### Verification
```
go build ./...
go test ./...
```
All tests must pass. Run the playground server and confirm no unexpected stdout
output during a render request.

---

## Phase 2 — Enforce clean API boundaries

### Goal
`pkg/nagare` must be the **sole** public entry point for rendering. `cmd/main.go`
must not import `pkg/diagram` directly. The chart-detection logic must not be
duplicated.

### Motivation
Currently `cmd/main.go` imports both `pkg/nagare` and `pkg/diagram`, so the
boundary between "public API" and "internal" is leaky. The `handleRender` legacy
endpoint also re-implements the chart-vs-diagram detection that already lives in
`nagare.RenderToSVG`.

### Changes

#### `pkg/nagare/nagare.go`
- Add a new exported function:
  ```go
  // RenderToHTML is the legacy endpoint helper used by the HTTP server.
  // For diagrams it returns the SVG; for charts it returns the full HTML page.
  func RenderToHTML(code string) (string, error)
  ```
  Move the chart-vs-diagram + `chart.RenderHTML()` logic out of `cmd/main.go`
  into this function so the cmd layer stays thin.
- Keep `RenderToSVG` as-is.
- Add `CreateDiagramWithSize` as a package-level function in `pkg/nagare` (delegating
  to `diagram.CreateDiagramWithSize`) so callers needing the canvas size do not have
  to import `pkg/diagram`.

#### `cmd/main.go`
- Remove the direct imports of `pkg/diagram` and `pkg/chart`.
- Rewrite `handleRender` to call `nagare.RenderToHTML(input)`.
- Rewrite `handleTest` to call `nagare.RenderToSVG(code)` (or the new helper).
- The file should import only `pkg/nagare` for rendering (plus stdlib packages).

#### `pkg/diagram/diagram.go`
- `CreateDiagram` and `CreateDiagramWithSize` remain exported for now (they are
  used by `pkg/nagare`), but add a package-level doc comment noting they are
  internal helpers and should not be called directly from outside the module.
- The long-term intent (achievable in a later phase) is to unexport them once all
  external callers have been routed through `pkg/nagare`.

### Verification
```
go build ./...
go test ./...
```
Confirm `cmd/main.go` no longer imports `pkg/diagram` or `pkg/chart`.

---

## Phase 3 — Rename "Legacy" to meaningful names

### Goal
Replace every use of `Legacy` as a type or function name with a name that describes
what the type actually does.

### Motivation
`Legacy` signals "this is temporary / bad". In production code it makes code hard
to read and debug, and misleads contributors into thinking the type will soon be
removed (when in fact it is the working implementation).

### Rename map

| Old name | New name | File |
|----------|----------|------|
| `browser.Legacy` | `browser.Component` | `pkg/diagram/components/browser/browser.go` |
| `browser.NewLegacy` | `browser.New` | same file |
| `vm.Legacy` | `vm.Component` | `pkg/diagram/components/vm/vm.go` |
| `vm.NewLegacy` | `vm.New` | same file |
| `server.Legacy` | `server.Component` | `pkg/diagram/components/server/server.go` |
| `server.NewLegacy` | `server.New` | same file |
| `terminal.Legacy` | `terminal.Component` | `pkg/diagram/components/terminal/terminal.go` |
| `terminal.NewLegacy` | `terminal.New` | same file |
| `database.Legacy` | `database.Component` | `pkg/diagram/components/database/database.go` |
| `database.NewLegacy` | `database.New` | same file |
| `cylinder.Legacy` | `cylinder.Component` | `pkg/diagram/components/cylinder/cylinder.go` |
| `cylinder.NewLegacy` | `cylinder.New` | same file |
| `led.Legacy` | `led.Component` | `pkg/diagram/components/led/led.go` |
| `led.NewLegacy` | `led.New` | same file |
| `messagequeue.Legacy` | `messagequeue.Component` | `pkg/diagram/components/messagequeue/messagequeue.go` |
| `messagequeue.NewLegacy` | `messagequeue.New` | same file |
| `cdn.Legacy` | `cdn.Component` | `pkg/diagram/components/cdn/cdn.go` |
| `cdn.NewLegacy` | `cdn.New` | same file |
| `apigateway.Legacy` | `apigateway.Component` | `pkg/diagram/components/apigateway/apigateway.go` |
| `apigateway.NewLegacy` | `apigateway.New` | same file |
| `backgroundworker.Legacy` | `backgroundworker.Component` | `pkg/diagram/components/backgroundworker/backgroundworker.go` |
| `backgroundworker.NewLegacy` | `backgroundworker.New` | same file |
| `packagecomponent.Legacy` | `packagecomponent.Component` | `pkg/diagram/components/packagecomponent/packagecomponent.go` |
| `packagecomponent.NewLegacy` | `packagecomponent.New` | same file |
| `artifact.Legacy` | `artifact.Component` | `pkg/diagram/components/artifact/artifact.go` |
| `artifact.NewLegacy` | `artifact.New` | same file |
| `rectangle.Legacy` | `rectangle.Component` | `pkg/diagram/components/rectangle/rectangle.go` |
| `rectangle.NewLegacy` | `rectangle.New` | same file |
| `core.LegacyDrawer` | `core.SVGDrawer` | `pkg/diagram/components/core/types.go` |
| `core.LegacyAdapter` | `core.DrawerAdapter` | same file |

### Callers to update
`pkg/layout/layout.go` holds type-assertions against every `*xxx.Legacy` type.
Update each `case *diagrambrowser.Legacy:` → `case *diagrambrowser.Component:` etc.
in `syncComponentGeometry`, `syncVMChildGeometry`, and `buildComponentTree`.

`pkg/diagram/diagram.go` — no direct use of `Legacy` types, but verify imports.

### Verification
```
go build ./...
go test ./...
```

---

## Phase 4 — Consolidate geometry types

### Goal
Have a single canonical `Shape` (a positioned rectangle) and a single `Point` type,
both living in `pkg/diagram/components/core`. Remove the redundant copies.

### Motivation
Currently:
- `components.Shape{X, Y, Width, Height, AlignmentRefs}` is used by layout and by
  every legacy component.
- `core.BoundingBox{X, Y, Width, Height}` is used by `core.Component` (the
  aspirational new interface).
- `layout.Point{X, Y}` and `core.Point{X, Y}` are identical structs in two files.

This forces every component to depend on `pkg/components` even if it does not use
templates. It also makes `core.Component` (the intended final interface) awkward
because it references a different bounding-box type than what the rest of the
pipeline uses.

### Approach

#### Step 1 — Promote `core.BoundingBox` → canonical `Shape`
In `pkg/diagram/components/core/types.go`, rename `BoundingBox` to `Shape` and add
the `AlignmentRefs` field that `components.Shape` currently has:
```go
type Shape struct {
    X, Y, Width, Height float64
    AlignmentRefs       map[string]string // deferred alignment resolution
}
```

#### Step 2 — Alias in `pkg/components`
In `pkg/components` (the SVG template layer), replace the struct definition with a
type alias pointing at the core type:
```go
import "github.com/saasuke-labs/nagare/pkg/diagram/components/core"
type Shape = core.Shape
```
This keeps all existing code compiling without changes.

#### Step 3 — Merge the `Point` types
Delete `layout.Point` from `layout.go` and replace it with `core.Point` everywhere
in that file. Update `layout.Arrow` to use `core.Point`.

#### Step 4 — Update `core.Component` interface
Change `BoundingBox()` to return `core.Shape` instead of `core.BoundingBox`:
```go
type Component interface {
    Draw() (string, error)
    Shape() core.Shape       // replaces BoundingBox()
    Port(name string) (core.Point, error)
}
```

#### Step 5 — Update `core.ShapeFromProps`
`core.ShapeFromProps` currently returns `components.Shape`. Change it to return
`core.Shape`. Since `components.Shape` is now a type alias, this is a no-op at the
call site.

### Verification
```
go build ./...
go test ./...
```
After this phase `layout.Point` no longer exists; `core.Shape` is the single
rectangle type used everywhere.

---

## Phase 5 — Dumb orchestrator: component self-registration

### Goal
Each component package owns its own default dimensions and build logic. The layout
engine becomes a generic loop over a registry rather than a hardcoded switch over
every known type.

### Motivation
`pkg/layout/layout.go` currently has four type-switch statements and fifteen
nearly-identical `buildXxx` functions. Adding a new component type requires editing
layout.go in at least four places. This makes the orchestrator too "smart" and
components too "dumb".

### Design

#### `pkg/layout/registry.go` (new file)
```go
package layout

// BuildFunc is the signature that each component must provide to integrate with layout.
type BuildFunc func(node parser.Node, container *ContainerSpec) (components.Component, error)

// Descriptor describes a component type to the layout engine.
type Descriptor struct {
    DefaultWidth  float64
    DefaultHeight float64
    Build         BuildFunc
}

var registry = map[string]Descriptor{}

// Register adds a component descriptor. Typically called from init() in each
// component package.
func Register(typeName string, d Descriptor) {
    registry[typeName] = d
}

// Lookup returns the descriptor for a component type, or a fallback.
func Lookup(typeName string) (Descriptor, bool) {
    d, ok := registry[typeName]
    return d, ok
}
```

`ContainerSpec` is a small struct that carries the absolute bounding box of the
parent VM (nil for top-level components):
```go
type ContainerSpec struct {
    AbsX, AbsY       float64 // absolute position of the content-area origin
    Width, Height    float64 // content-area size
}
```

#### One registration per component package
Each `pkg/diagram/components/xxx/xxx.go` adds an `init()` function:
```go
func init() {
    layout.Register("Server", layout.Descriptor{
        DefaultWidth:  200,
        DefaultHeight: 140,
        Build:         buildFromNode,
    })
}

func buildFromNode(node parser.Node, container *layout.ContainerSpec) (components.Component, error) {
    comp := New(node.Text)
    // apply state props, set shape …
    return comp, nil
}
```

> **Note**: The import cycle risk is real here.
> `pkg/layout` currently imports all component packages. With the registry pattern
> the dependency must be reversed: each component package imports `pkg/layout` (for
> `Register`), and `pkg/layout` does **not** import any component package.
>
> To achieve this without a circular import, extract the `Register`/`Lookup` registry
> into its own tiny package `pkg/layout/registry` (or `pkg/compregistry`) that has no
> imports from the rest of the pipeline. `pkg/layout` imports `pkg/compregistry`;
> each component package also imports `pkg/compregistry`. The component packages
> still import `pkg/components` for templates.

#### Simplify `buildComponentTree`
Replace the full switch:
```go
func buildComponentTree(node parser.Node, nodeIndex map[string]components.Shape) []components.Component {
    d, ok := registry.Lookup(string(node.Type))
    if !ok {
        d, _ = registry.Lookup("Rectangle") // fallback
    }
    comp, err := d.Build(node, nil)
    if err != nil {
        return nil
    }
    nodeIndex[node.Text] = comp.Shape()
    return []components.Component{comp}
}
```

#### Simplify `layoutVMChildren`
```go
func layoutVMChildren(parent parser.Node, vm *vm.Component, nodeIndex map[string]components.Shape) {
    container := containerSpec(vm)
    for _, child := range parent.Children {
        d, ok := registry.Lookup(string(child.Type))
        if !ok {
            d, _ = registry.Lookup("Rectangle")
        }
        comp, err := d.Build(child, container)
        if err != nil {
            continue
        }
        nodeIndex[child.Text] = comp.Shape()
        vm.AddChild(comp)
    }
}
```

#### Simplify `syncComponentGeometry`
Once every component implements `core.Component` (i.e., has `Shape()` and can
`SetShape()`), the sync becomes a simple loop with no type assertions:
```go
func syncComponentGeometry(children []components.Component, nodeIndex map[string]components.Shape) {
    for _, child := range children {
        if setter, ok := child.(ShapeSetter); ok {
            if resolved, found := nodeIndex[child.ID()]; found {
                setter.SetShape(resolved)
            }
        }
    }
}
```
Add a `ShapeSetter` interface and an `ID()` method to `core.Component`:
```go
type Component interface {
    ID() string
    Draw() (string, error)
    Shape() core.Shape
    SetShape(core.Shape)
    Port(name string) (core.Point, error)
}
```

#### Simplify `diagram.drawNode`
In `pkg/diagram/diagram.go`, the `drawNode` switch can be replaced by a single call:
```go
func (d *Diagram) drawNode(node *RenderNode, shape core.Shape) string {
    comp, err := registry.Build(node.Type, node.ID, node.Props)
    if err != nil {
        return fmt.Sprintf("<!-- unknown component %q -->", node.Type)
    }
    svg, _ := comp.Draw()
    return svg
}
```
This requires each component's `Build` function to also accept `(id string, props map[string]any)`.

### Migration strategy
Do this in two sub-steps:
1. Add the registry and `init()` registrations while keeping the existing switches as
   fallbacks. Verify tests pass.
2. Replace each switch arm with a registry call. Delete the old switch. Verify tests pass.

### Verification
```
go build ./...
go test ./...
```
After this phase, `layout.go` should have no imports of individual component packages,
and every `buildXxx` function should be gone.

---

## Phase 6 — Deduplicate shared utilities

### Goal
Ensure each utility function has exactly one implementation.

### Duplicates to eliminate

#### `floatProp` / `core.FloatProp`
- `pkg/diagram/diagram.go` defines `floatProp(props map[string]any, key string) float64`
- `pkg/diagram/components/core/render.go` defines `FloatProp(props map[string]any, key string, fallback float64) float64`
- **Action**: Delete `floatProp` from `diagram.go`; replace all call sites with
  `core.FloatProp(props, key, 0)`.

#### `actionMapFromAny`
- Identical implementation in both `pkg/diagram/diagram.go` and
  `pkg/diagram/components/database/database.go`.
- **Action**: Keep exactly one copy in `pkg/diagram/components/core/actions.go`
  (new small file), export it as `core.ActionMapFromAny`. Delete the duplicates.
  Update call sites in `database.go` and `diagram.go`.

#### `splitProps` / `props.ParseProps`
- `pkg/diagram/diagram.go` has `splitProps` and `parsePropsDefSafe` that partially
  re-implement what `pkg/props/props.go` does.
- **Action**: Add a new function to `pkg/props`:
  ```go
  // ParseToMap parses a raw props string into a map[string]any,
  // coercing numeric values.
  func ParseToMap(raw string) map[string]any
  ```
  Then delete `splitProps`, `parsePropsDefSafe`, and `coerceValue` from `diagram.go`;
  replace all call sites with `props.ParseToMap(raw)`.

#### `serializePropValue` / `rawPropsFromNode`
- These are only used internally in `diagram.go` and are not duplicated, but they
  are complex and fragile. Leave them as-is for now; they can be addressed in
  Phase 8 when `map[string]any` props are replaced by typed structs.

### Verification
```
go build ./...
go test ./...
```

---

## Phase 7 — Proper error propagation

### Goal
Replace every silent failure with a real `error` return. The pipeline should
surface errors to callers rather than silently producing empty or partial SVG.

### Motivation
Currently:
- `_ = comp.Props.Parse(raw)` — parse failures are silently discarded.
- `fmt.Printf("failed to parse …")` (replaced by no-ops in Phase 1) — failures
  disappear without reaching the caller.
- `layout.Calculate` returns `Layout` with no error path; parse failures simply
  produce zero-value geometry.
- `components/browser.go Draw()` returns `""` on template error (instead of an SVG
  comment or propagating up).

### Changes

#### `pkg/layout/layout.go` — add error returns
Change:
```go
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) Layout
```
to:
```go
func Calculate(node parser.Node, canvasWidth, canvasHeight float64) (Layout, error)
```
Collect non-fatal errors (e.g. unknown component types, failed prop parses) into a
`[]error` slice; decide whether to return the first error or a joined error. For
most parse failures it is reasonable to continue with defaults and return a warning.

Update `pkg/diagram/diagram.go → ParseDiagram` to handle the new signature.

#### `pkg/props/props.go` — return on unknown key
Currently the function logs a warning and continues when a prop key has no
matching struct field. This is acceptable behaviour but the log is gone after
Phase 1. Add a `[]string` return for unrecognised keys so callers can decide what
to do.

#### `pkg/diagram/components/*` — propagate `Draw() (string, error)`
Each component's `Draw()` currently returns only `string`. The `core.Component`
interface already defines `Draw() (string, error)`. Update the concrete
implementations in `pkg/components/*.go` to return `(string, error)` so they can
be used directly as `core.Component` without the `LegacyAdapter` wrapper.

#### `pkg/nagare/nagare.go`
`RenderToSVG` already returns `(string, error)`; ensure all internal error paths
now reach this return value instead of being swallowed.

### Migration note
This phase has the most surface area. Work file by file, running `go build ./...`
after each change. The geometry consolidation (Phase 4) and component registration (Phase 5)
should already have made most error paths explicit by this point.

### Verification
```
go build ./...
go test ./...
```
Introduce a deliberately malformed `.nagare` snippet in a test and assert that
`nagare.RenderToSVG` returns a non-nil error rather than empty SVG.

---

## Phase 8 — Reduce `map[string]any` props surface

### Goal
Replace the `map[string]any` prop bag that flows through the render pipeline
(`RenderNode.Props`, `DrawFromRenderNode` arguments) with typed structs or a
clearly-bounded accessor API. This is an aspirational long-term improvement and
can be done incrementally.

### Motivation
`map[string]any` provides no compile-time type safety. Every component must
defensively type-assert every value it reads. The `FloatProp` / `actionMapFromAny`
helpers exist precisely to hide this pain, which is a symptom of the underlying
type weakness.

### Approach (incremental)

#### Step 1 — Introduce `RenderProps` value type
```go
// pkg/diagram/renderprops/renderprops.go
package renderprops

type RenderProps struct {
    X, Y, W, H  float64
    RawProps     string
    Actions      map[string][]map[string]any
    Extra        map[string]any // remaining typed-unknown props
}

func FromMap(m map[string]any) RenderProps { … }
func (p RenderProps) ToMap() map[string]any { … } // backward-compat bridge
```

#### Step 2 — Thread `RenderProps` through `DrawFromRenderNode`
Change the signature of `DrawFromRenderNode` in all component packages from:
```go
func DrawFromRenderNode(id string, nodeProps map[string]any) string
```
to:
```go
func DrawFromRenderNode(id string, props renderprops.RenderProps) string
```
The callsite in `diagram.go → drawNode` calls `renderprops.FromMap(localProps)`
before dispatching.

#### Step 3 — Type-safe component Props structs
Each component already has a typed `Props` struct (e.g. `browser.BrowserProps`).
The `DrawFromRenderNode` function already calls `comp.Props.Parse(raw)`. Once
`RenderProps` carries `RawProps` directly, the parse becomes cleaner. Future work
can skip the round-trip through raw strings entirely.

### Note
This phase may be deferred until after Phases 1–7 are complete and the team is
comfortable with the new structure. The `map[string]any` interface works correctly;
this phase improves maintainability rather than fixing bugs.

### Verification
```
go build ./...
go test ./...
```

---

## Execution order

| Phase | Depends on | Risk | LOC delta |
|-------|-----------|------|-----------|
| 1 — Debug output | — | very low | −60 |
| 2 — API boundaries | 1 | low | ±30 |
| 3 — Rename "Legacy" | 1 | low | ±0 (pure rename) |
| 4 — Consolidate geometry types | 3 | medium | −40 |
| 5 — Component self-registration | 3, 4 | high | −500 in layout.go |
| 6 — Deduplicate utilities | 1, 2 | low | −80 |
| 7 — Error propagation | 1, 5, 6 | medium | +100 |
| 8 — Typed props | 4, 5, 7 | medium | +200 |

Phases 1, 2, 3 can be done in any order (or in parallel) since they have no
semantic dependencies on each other.

Phase 5 is the largest change. Validate it heavily with integration tests before
proceeding to Phase 7.

---

## What stays the same

- The DSL syntax is unchanged.
- The public `pkg/nagare` API surface (`RenderToSVG`, `RenderFileToFile`) is
  unchanged.
- SVG template files under `pkg/components/templates/` are unchanged.
- `pkg/tokenizer` and `pkg/parser` are unchanged.
- The HTTP server endpoints in `cmd/main.go` stay the same from the caller's
  perspective; only the internal implementation changes.
- The chart sub-system (`pkg/chart`) is unchanged throughout all phases.
