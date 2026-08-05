# Verification Checklist — Zoom provider-owned inventory parity, R1

## Wave 0

- [x] Report ledger body parses with `jq empty`.
- [x] Report ledger verifies 1,913 unique method/path rows with 3 covered, 1,839 implementable-now,
  17 provider-restricted, and 54 deprecated rows.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/zoom` passes.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1` passes.
- [x] Commit `46eff2585` contains only `internal/connectors/defs/zoom/api_surface.json`.

The 54 report rows using legacy `excluded.category=deprecated` were normalized mechanically to the
accepted operation-ledger-v1 representation (`model=deprecated`, blocked by default, retained
provider reason and source URL). This was a validator format requirement, not a reclassification.

## Wave 1

- [x] A Zoom-local command-surface test was observed red before `cli_surface.json` existed:
  `CommandSurface()` was nil and the existing stream rows were unreachable as `pm zoom` commands.
- [x] `go test ./internal/connectors/defs/zoom -count=1` passes. It covers the inventory arithmetic,
  three exact command-to-stream/API bindings, real command-runner preflight, and fixture-backed
  execution with a bounded `--limit` and `--user-id` override.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/zoom` passes.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs/zoom --check` passes with no
  derived-field changes.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1` passes.
- [x] `go test ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1` passes.
- [x] `go test ./internal/cli -count=1` passes.
- [x] `go vet ./internal/connectors/defs/zoom` and `go vet ./internal/cli` pass; targeted
  `golangci-lint run ./internal/connectors/defs/zoom` reports zero issues.
- [x] `go build ./cmd/pm` passes. The built `./pm` returns zero for `pm zoom`, `pm help zoom`,
  `pm zoom users list --help`, `pm zoom meetings list --help`, and `pm zoom webinars list --help`;
  those help/preflight paths do not resolve credentials or call Zoom.
- [x] `pm docs generate --dir docs/cli` regenerated `docs/connectors/zoom/{MANUAL,SKILL}.md`.
  `make docs-check-no-build` passes.
- [x] `pnpm run gen:website-data`, `pnpm run typecheck`, and `pnpm run test:scripts` pass; generated
  website data exposes the Zoom command surface and its honest inventory counts.
- [x] `make connector-boundary` and `git diff --check` pass.
- [x] Manual code review against the generated GSD code-review prompt found no connector-local
  correctness, safety, scope, or generated-artifact drift issue. The registered GSD phase is absent
  from the shared roadmap, so its worker workflow cannot be started for this isolated task.

## Scope and handoff

- Changed product files are Zoom-local plus its generated manuals and website catalog output.
- No live provider request, credentials, provider write, shared-runtime change, or redaction policy
  was used.
- The full repository suite and `make verify` are intentionally left to CI/no-mistakes under the
  repository timeout guidance. Per the task gate, this branch stops after the Wave 1 commit; do not
  start no-mistakes or create a PR until firstmate instructs it.
