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

## Foundation re-gate

- Re-gate scope: the merged non-redacting output-policy, typed multipart, `rest_write`, destructive
  approval, API-surface provenance, and rate-limit foundations were reviewed without changing a
  Zoom ledger disposition or reason.
- Tool / result: `go run ./cmd/pm connectors inspect zoom --json` reported `write=false` and the
  three declared streams; `go run ./cmd/pm zoom --help` exposed only the Wave 1 users, meetings,
  and webinars command groups. The executable endpoint count remains **3 → 3**, with **0 → 0**
  operation-row promotions and three implemented CLI commands.
- The 1,910 remaining endpoint rows have no complete Zoom-local typed operation contract plus
  executable command surface, so they remain blocked. No `operations.json`, Zoom write action,
  credential, live request, redaction policy, or Wave 2 route was added.
- Focused proof: `go test ./internal/connectors/defs/zoom -run
  'TestProviderInventoryLedgerIsComplete|TestCoveredStreamsHaveReachableCommands' -count=1`
  passes, retaining the 1,913-row ledger and exactly the three reachable command bindings.
