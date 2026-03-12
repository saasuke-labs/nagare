# RFC: Action-Driven Arrow Creation and Centralized Arrow Resolution

- **Status**: Draft
- **Target release**: TBD
- **Author**: Codex (proposal)
- **Last updated**: 2026-03-12

## 1. Summary

This RFC proposes adding **action-driven arrow creation** and **centralizing arrow endpoint/routing logic** in shared parser/layout resolver flow, so diagram components perform less bespoke connection math while the Arrow component remains action-agnostic.

Nagare already supports explicit connection ports (including `n`, `ne`, `e`, `se`, `s`, `sw`, `w`, `nw`) in direct arrow syntax. This RFC does **not** redesign that capability; it builds on top of it.

Primary new capability:

```nagare
@b.request(target: @s, dir: lr)
```

which should generate the same connection result as explicit arrow syntax by resolving endpoints through the shared arrow resolution path.

## 2. Motivation

Current connection behavior works, but connection intent can be spread across component-specific logic. We want to:

1. Let component actions create arrows directly (`@component.action(...)`).
2. Ensure action-generated arrows and explicit arrows use one normalized connection pipeline.
3. Move endpoint/route calculation out of individual diagram components and into centralized arrow resolution.
4. Preserve existing explicit port syntax and behavior.

This keeps `pkg/diagram/components/*` focused on rendering and local component behavior, while parser/layout own connection semantics.

## 3. Non-goals

- Replacing or deprecating existing explicit arrow syntax.
- Reworking SVG output scope (SVG remains the output target).
- Building a full obstacle-avoidance auto-router in this phase.

## 4. Existing behavior to preserve

### 4.1 Explicit ports are already supported

The direct-connection form remains valid and unchanged:

```nagare
@b.e --> @s.w
@a.ne --> @b.sw
```

Supported compass ports: `n`, `ne`, `e`, `se`, `s`, `sw`, `w`, `nw`.

### 4.2 Backward compatibility

All current connection parsing/routing behavior for existing diagrams must remain compatible.

## 5. Proposed additions

### 5.1 Action-generated connection intents

Allow specific component actions to emit connection intents, for example:

```nagare
@b.request(target: @s, dir: lr)
```

Expected semantics:

- `target` identifies the destination component.
- `dir` is a direction hint.
- When `dir` is present and explicit ports are absent, the resolver infers source/target ports.

Initial direction aliases:

- `lr` → source `w`, target `e` (user-requested mapping)
- `rl` → source `e`, target `w`
- `tb` → source `s`, target `n`
- `bt` → source `n`, target `s`

### 5.2 Centralized arrow resolution contract

Introduce a normalized connection-intent model (from both explicit arrows and action-generated arrows):

- `FromID`
- `ToID`
- `FromPort` (optional)
- `ToPort` (optional)
- `DirectionHint` (optional)
- `ActionName` (optional)
- style/animation metadata

One resolver should convert intent → concrete arrow geometry.

### 5.3 Component boundary responsibilities

- Components may expose supported action aliases at the component boundary (e.g., `Browser.request`, `Browser.response`).
- The component owning the action is responsible for mapping that action into an Arrow instance input (for example: `request` -> solid line, `response` -> dashed line).
- Components should avoid implementing custom per-action endpoint math.
- Arrow placement/routing decisions should flow through shared resolver logic.

## 6. Pipeline impact

### 6.1 Tokenizer

Likely minimal/no changes. Action arguments like `target: @s` and `dir: lr` must be preserved accurately.

### 6.2 Parser (`pkg/parser`)

Enhance parser to:

1. Keep existing explicit connection parsing intact.
2. Recognize action forms that should produce connection intents.
3. Normalize explicit and action forms into the same connection-intent shape.
4. Emit clear errors for invalid action connection parameters (missing `target`, unknown `dir`, etc.).

### 6.3 Layout (`pkg/layout`)

Enhance connection resolution stage to:

1. Resolve component instances by ID.
2. Resolve explicit ports when provided.
3. Infer ports from direction hint when explicit ports are absent.
4. Route arrow points via shared helpers (e.g., `routeArrowPoints`) and output concrete geometry for rendering.

### 6.4 Diagram components (`pkg/diagram/components/*`)

Simplify diagram components by removing duplicated connection calculations.

Keep component packages focused on:

- component render-node conversion,
- local rendering semantics,
- optional action alias declarations.

### 6.5 Arrow component (`pkg/components/arrow.go` + diagram adapter)

Arrow consumes resolved endpoint/routing metadata and visual style props only (instead of requiring each component to precompute custom anchor math):

- resolved source point/port
- resolved target point/port
- optional routed waypoints
- style metadata (solid/dashed, marker options, etc.)

Arrow does **not** know or branch on other component action names; action semantics are resolved before Arrow is instantiated.

## 7. Implementation plan (phased)

### Phase 1 — Parser/domain normalization

- Add/extend normalized connection-intent types.
- Parse action-generated connection intents from `@id.action(...)` forms.
- Preserve existing explicit arrow parsing unchanged.
- Add parser tests covering action forms + validation errors.

### Phase 2 — Shared resolver in layout

- Implement intent resolver for explicit + action-generated intents.
- Add direction-to-port mapping (`lr`, `rl`, `tb`, `bt`).
- Ensure compass ports (`n`, `ne`, `e`, `se`, `s`, `sw`, `w`, `nw`) remain supported.
- Add layout tests for inferred ports and fallback behavior.

### Phase 3 — Arrow integration and component simplification

- Ensure arrow render path accepts centralized resolver output and style props provided by action owners.
- Remove duplicated component-side endpoint logic where redundant.
- Add integration/golden tests for action-generated and explicit connections (including style parity expectations such as request=solid/response=dashed).

### Phase 4 — Docs and examples

- Update README with action-generated arrow examples.
- Add/refresh playground examples showing explicit vs action forms.
- Update AGENTS notes if resolver conventions evolve.

## 8. Validation strategy

1. **Parser tests**: action forms, target parsing, direction validation.
2. **Layout tests**: intent resolution, direction mapping, explicit-port compatibility.
3. **Renderer/diagram tests**: SVG output parity between explicit and action-generated arrows.
4. **Regression checks**: existing fixture diagrams render unchanged.

## 9. Risks and mitigations

- **Risk: action semantics diverge from explicit arrow behavior**  
  **Mitigation**: one normalized intent resolver + parity tests.

- **Risk: unclear direction alias expectations (`lr`)**  
  **Mitigation**: lock mapping in docs/tests (`lr` = source `w`, target `e`).

- **Risk: component logic remains fragmented**  
  **Mitigation**: migrate endpoint math to central resolver; keep component responsibilities narrow.

## 10. Open questions

1. Should action-generated arrows support multiple targets in V1?
2. Should `dir` also support verbose aliases (`left-right`) in addition to short forms?
3. On invalid/missing ports, should behavior fail hard or fall back to a default policy?

## 11. Documentation updates after implementation

When implementation lands:

- Update `README.md` examples to include action-generated arrows.
- Add/refresh examples under `static/examples/`.
- Record resolver/action conventions in `AGENTS.md` for future contributors and agents.
