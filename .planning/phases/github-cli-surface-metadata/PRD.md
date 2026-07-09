# PRD: GitHub CLI Surface Metadata

## Objective

Implement issue #34: promote the GitHub connector CLI surface from research into validated,
docs-only production metadata.

## Scope

- Add optional `cli_surface.json` loader support for connector bundles.
- Add `connectorgen validate` checks for command references.
- Add a GitHub `internal/connectors/defs/github/cli_surface.json` file.
- Keep runtime command dispatch out of scope.
- Keep raw API and direct-write execution out of scope.

## Non-goals

- No `pm github ...` dynamic command dispatch in this slice.
- No GraphQL executor in this slice.
- No reverse ETL execution outside plan, preview, approval, execute.
- No generic shell, generic SQL write, or unrestricted generic HTTP write.

## Acceptance Criteria

- `go test ./internal/connectors/engine -run CLISurface` passes.
- `go test ./cmd/connectorgen -run CLISurface` passes.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passes.
- Implemented CLI commands in `cli_surface.json` resolve to existing streams or writes.
- Unsupported local workflow commands remain visible as metadata and blocked from dispatch.
- Examples contain no secret-looking values.

## Manual GSD Fallback

`scripts/programming-loop.mjs` and `scripts/tdd-gate.mjs` are not present in this worktree. This
phase uses the manual GSD loop required by `AGENTS.md`: plan, red tests, green implementation,
refactor, verification, and recorded evidence.
