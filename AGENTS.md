# Agent Notes for the Repository

## Active Areas
- Prioritize changes in the main Go module under `nagare/`.
- The `.github/` workflows are live and should be updated carefully when necessary.
- Everything else in the repo (e.g., `motion/`, `server/`, `nagare-go/`) are PoCs slated for removal—avoid modifying them unless explicitly requested.

## Project Overview
- Nagare is a text-to-diagram renderer in the spirit of Mermaid or PlantUML.
- The DSL should remain as simple as possible while retaining essential expressiveness.
- Layout is currently explicit; auto-layout may arrive later, so design syntax and data structures with that future flexibility in mind.
- Upcoming milestones include adding animation directives (fade-in/out, animated WebP export). Keep this roadmap in mind when shaping APIs and formats.

## Contribution Guidelines
- Follow the pipeline architecture documented in `nagare/AGENTS.md` when working inside the Go module.
- Prefer explicit declarations in the DSL over hidden defaults until auto-layout/animation features are ready.
- Update the README with new syntax examples or behavior changes that affect end users.
- Run `go test ./...` within `nagare/` before submitting Go changes.

## PR Expectations
- Summaries should call out how the change advances the static rendering pipeline or prepares for future animation support.
- Link to preview artifacts (e.g., rendered SVGs) when available.
