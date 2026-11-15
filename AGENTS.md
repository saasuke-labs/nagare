# Nagare — agent guide

Nagare renders architecture diagrams and charts to SVG from a small text DSL. One Go binary
that is a CLI, an HTTP server with a browser playground, and an importable Go library.
Module `github.com/saasuke-labs/nagare`, Go 1.24, **zero external dependencies** (no `go.sum`;
keep it that way unless there is a strong reason).

Output is SVG only. WebP/PNG export is intentionally out of scope and handled externally;
do not add image conversion to the runtime or to CI.

## Where it sits in Saasuke Labs

| Repo | What it is | Relation to nagare |
| --- | --- | --- |
| `saasuke-labs/gengo` | Static site generator (Go) | Imports `pkg/nagare` as a library and renders `nagare` code blocks in Markdown. A change to the public API or to SVG output reaches gengo when it bumps the nagare version in its `go.mod`. |
| `saasuke-labs/kotomi` | Comments and reactions server for static sites | No code dependency. |
| `saasuke-labs/site-source` | Source of saasuke-labs.com (builds into `saasuke-labs.github.io`) | The release workflow sends it a `product-release` dispatch so the releases page updates. |

`docs/index.html` is the product page served at saasuke-labs.com/nagare/. It is hand-written
HTML that loads its CSS, JS and logo from the main site (`/static/site.css`, `/static/site.js`),
so it only looks right when served under that domain. Diagrams in it are pasted SVG output:
when rendering changes, re-render the examples and replace them.

## Commands

```bash
go test ./...                                  # all tests (fast, no network)
go test ./pkg/layout -run TestName             # one test
gofmt -l .                                     # must print nothing
go vet ./...
go run ./cmd -input in.nagare -output out.svg  # CLI mode
go run ./cmd -port 8080                        # server + playground at /
go build -o nagare ./cmd                       # overwrites the tracked binary, see Gotchas
```

A `justfile` wraps the same commands (`just test`, `just start`, `just gen in.nagare`,
`just start-watch` for hot reload with `air`).

- No workflow runs `go test`. CI only builds the Docker image and renders the preview
  diagrams, so run tests, `gofmt` and `go vet` locally before pushing.
- Run the server from the repo root: the playground is read from the relative path
  `static/playground.html`.
- The CLI needs both `-input` and `-output`; with neither it starts the server.

## Architecture

Entry points: `cmd/main.go` (flags, CLI mode, HTTP handlers) and `pkg/nagare` (the public
library API: `RenderToSVG`, `RenderToHTML`, `CreateDiagramWithSize`, `RenderFileToFile`).
Everything else is internal in practice; outside callers should only import `pkg/nagare`.

`RenderToSVG` picks one of two independent pipelines:

**Charts** — input that starts with `chart` goes to `pkg/chart` (`chart.Parse` →
`Chart.RenderSVG`). It has its own line-based parser and renderer and shares nothing with
diagrams. Supports line/dashed/dotted/bar/marker series, stacked bars, multiple y-axes and
date or duration values.

**Diagrams** — everything else goes through `pkg/diagram.CreateDiagram`:

1. `pkg/tokenizer` — DSL text → tokens.
2. `pkg/parser` — tokens → `parser.Node` AST. `@name(...)` blocks are stored as states
   (`Globals` by name, `GlobalStates` in source order), `a.e --> b.w` as `Connections`.
   Nesting deeper than one level is a parse error.
3. `pkg/layout` — `layout.Calculate` instantiates components, resolves geometry (`x/y/w/h`,
   alignment references like `y:&browser.c`), fills `NodeIndex` (id → absolute shape) and
   routes arrows, both explicit connections and action-generated ones.
4. `pkg/diagram` — builds a `RenderNode` tree (props merged in order: inline props, then
   `@id(...)` state, then named `@state`), walks it wrapping each node in
   `<g transform="translate(...)">`, and calls the component's `DrawFromRenderNode`. Arrows
   are drawn last from `Layout.Children`. The canvas is transparent.

Packages:

- `pkg/diagram/components/<type>/` — one package per component: props struct with `prop:"..."`
  tags, defaults, `DrawFromRenderNode(id, props)`.
- `pkg/diagram/components/core` — shared types (`Shape`, `Point`, ports, `ConnectionIntent`,
  `ActionConnector`, prop helpers).
- `pkg/diagram/components/style` — the shared visual system: `Light`/`Dark` themes, accent
  palette, icons, text fitting and the `Card`/`Window` primitives every component is drawn with.
- `pkg/components` — older component structs and `html/template` SVG templates. Still used
  for `Arrow`, `Shape` and the `Component` interface, and by the adapters in layout.
- `pkg/props` — prop string parsing (`ParseToMap`, `ParseProps` into tagged structs).
- `pkg/renderer` — legacy; nothing imports it. Do not extend it.
- `pkg/version` — set by goreleaser ldflags.

### DSL in one screen

```text
@layout(w: 950, h: 400, theme: dark)

browser:Browser(url: "https://example.com", x: 50, y: 100, w: 200, h: 150)
vps:VM(title: "home@ubuntu", x: 300, y: &browser.c, w: 600, h: 300) {
    nginx:Server(title: "nginx", port: 80, x: 50, y: 20, w: 200, h: 50)
}
db:Database
@db(x: 300, y: 200, w: 400, h: 400)

browser.e --> nginx.w
@db.read(begin: 1s, dur: 1s)
```

- `@layout(...)` sets canvas size and theme (`light` is the default).
- `id:Type(props)` declares a component; `{ ... }` nests children (one level).
- `@id(props)` applies props by id; `id:Type@state` plus `@state(props)` applies a named state.
- `a.e --> b.w` is an explicit arrow between ports (`n`, `s`, `e`, `w`, `c`).
- `@id.action(props)` is an action: an animation (`db.read`) or an action-generated arrow
  (`browser.request`).
- `&id.port` references another component's geometry.
- There is no comment syntax. Stray words are parsed as nodes of an unknown type.

Types: `Browser`, `VM`, `Server`, `Terminal`, `Database`, `Cylinder`, `Led`,
`MessageQueue`/`Queue`, `CDN`/`Edge`, `APIGateway`, `BackgroundWorker`, `Package`,
`Artifact`/`File`, `Rectangle`. Colour props `bg`, `fg`, `accent`, `contentBg` override the
theme. The README has the full syntax and chart reference.

## Adding or changing a component

A component type is registered by hand in three places. Missing one fails silently:

1. `buildComponentTree` in `pkg/layout/layout.go` (top-level), plus a `build<Type>` helper
   and default size constants.
2. `layoutVMChildren` in the same file, if it may live inside a `VM`.
3. `drawNode` in `pkg/diagram/diagram.go`.

An unknown type gets rectangle geometry in layout but draws nothing, because `drawNode`
returns an empty string for it. Arrows to it still render.

Then add a `.nagare` file under `.github/testdiagrams/` so the PR preview shows it, and
update `README.md`, `docs/index.html` and the examples in `static/playground.html` if the
syntax changed.

## Conventions

- Extend the existing stages (tokenizer → parser → layout → diagram) instead of skipping
  around them.
- Reuse the helpers in `pkg/props`; do not parse props inside components.
- Draw with the `style` package (`Card`, `Window`, theme tokens, accents) so components
  share one look in both themes. Do not hardcode colours a theme already provides.
- Components draw at local origin. `drawNode` sets `x`/`y` to 0 and the surrounding `<g>`
  does the placement; using absolute coordinates doubles the offset.
- Prefer composing from primitives (`Database` = `Cylinder` + two `Led`s) with
  parent-relative, proportional child geometry so compositions scale with the parent.
- Map alias actions at the parent boundary (`Database.read/write` → child `Led.blink`) so
  primitives keep a small action surface.
- The parser stays action-name agnostic. A component whose actions create arrows implements
  `core.ActionConnector` and returns `ConnectionIntent`s (for example `Browser.request` solid,
  `response` dashed); layout resolves coordinates. `Arrow` stays action-agnostic. Background:
  `docs/RFC_PORT_AWARE_ARROWS.md`.
- Keep per-component render-node conversion in the component package so
  `pkg/diagram/diagram.go` stays orchestration only.
- In layout, use `applyIDStateProperties`, `applyNamedStateProperties` and
  `routeArrowPoints` for new component types or connection rules.
- Inline props (`name:Type(x: ...)`), `@id(...)` and `name:Type@state` are all supported;
  prefer inline props in examples. Anonymous `Type(x: ...)` gets a generated id.

## Gotchas

- Arrow marker ids come from a global counter. Tests that compare SVG must call
  `components.ResetArrowMarkerCounter()` first.
- `pkg/diagram/diagram_test.go` compares against golden files in `pkg/diagram/fixtures/`.
  There is no update flag: when output changes on purpose, regenerate the SVG and replace
  the fixture by hand.
- Rendering prints debug lines to stdout (`fmt.Print*` calls in `pkg/parser/parser.go` and
  `pkg/components`). Library users such as gengo get them too. The CLI writes SVG to `-output`, so
  files stay clean.
- A built `nagare` binary is tracked in git even though `.gitignore` lists it. Do not commit
  a rebuilt one: `git checkout nagare` after `go build` or `just build`.
- Root-level `example.json`, `example2.json`, `scene.json`, `sample.mmd.json` and
  `catalog.svg` are leftovers that nothing references.
- `IMPROVEMENT_PLAN.md`, `REFACTOR_SUMMARY.md` and `CHART_IMPLEMENTATION.md` are historical
  write-ups, not current plans. Trust the code and this file over them.

## CI, release and deploy

- `render-preview.yml` (every PR): builds the CLI, renders every
  `.github/testdiagrams/*.nagare`, commits the SVGs to `.github/pr-images/pr-<n>/` on the PR
  branch and comments them on the PR. Expect a bot commit on your branch; pull before
  pushing again. Never edit `.github/pr-images/` by hand.
- `docker-build.yml` (every PR): `docker build .` must succeed. The image is `scratch` with
  the static binary and `static/`.
- `deploy-gcp.yml` (push to `main`): deploys the playground to Cloud Run through GitHub OIDC
  and Workload Identity Federation. Keep `docs/GCP_CLOUD_RUN_CI_SETUP.md` in sync when auth
  or deploy settings change. If auth fails with `iam.serviceAccounts.getAccessToken`, the
  deployer service account needs both `roles/iam.workloadIdentityUser` and
  `roles/iam.serviceAccountTokenCreator` for the GitHub `principalSet`.
- `release.yaml`: pushing a `v*.*.*` tag runs GoReleaser (linux/darwin/windows, amd64/arm64)
  and then notifies `site-source`. It can also be run manually for an existing tag.
  `install/install.sh` installs the latest release binary.

## More detail

- `README.md` — full DSL and chart syntax, playground features.
- `docs/LIBRARY_USAGE.md` — using nagare as a Go library.
- `docs/RFC_PORT_AWARE_ARROWS.md` — action-driven arrows design.
- `docs/GCP_CLOUD_RUN_CI_SETUP.md` — Cloud Run deployment setup.
