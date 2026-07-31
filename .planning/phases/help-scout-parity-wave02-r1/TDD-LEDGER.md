# Help Scout parity wave02 r1 — TDD ledger

## Red tests / validation targets

- [x] Added Help Scout-owned validation test before production bundle rewrite.
  - Red evidence: `go test ./cmd/connectorgen -run 'TestHelpScoutFullSurface' -count=1` failed because current bundle had no `writes.json` and only the legacy 4-stream surface.
- [x] Bundle validation target established before rewrite.
  - The repo-local `connectorgen validate` command validates a defs root, not a single connector dir; focused validation uses a temporary defs root containing only `help-scout`.

## Green targets

- [x] `api_surface.json` has 144 canonical operations and executable/blocked split: stream=45, write=65, operation=34.
- [x] `streams.json` declares 45 Help Scout read streams with schemas.
- [x] `writes.json` declares 65 reverse ETL actions; every action has `confirm: "destructive"`, and every DELETE has idempotent 404 handling plus redacted path identifiers.
- [x] `operations.json`/`cli_surface.json` represent blocked direct/report query reads and binary attachment downloads without enabling raw generic execution.
- [x] Docs and fixtures are sanitized and mention no live credentials/certification.
- [x] Parent/subissue captain-policy addendum appended idempotently to #212-#219 with `gh-axi`.

## Refactor / hardening targets

- [x] Kept changes connector-local plus Help Scout-owned test/planning artifacts.
- [x] Recorded foundation blockers rather than editing shared engine/CLI/runtime code.
- [x] Preserved issue body count tables; addendum appended after existing bodies.

## Evidence log

- GSD adapter fallback recorded in `PLAN.md` because `scripts/gsd prompt programming-loop ...` is unavailable in this checkout.
- Focused green evidence: `go test ./cmd/connectorgen -run 'TestHelpScoutFullSurface' -count=1`, focused temp-root `connectorgen validate`, `go test ./internal/connectors/conformance -run 'TestConformance/help-scout' -count=1`.
- Broader green evidence: `go vet ./...`, `go build ./cmd/pm`, `make connector-boundary`, `make connectorgen-validate`, `make smoke-no-build`, `make docs-check`, `make lint`, `git diff --check`.
- Incomplete evidence: `go test ./...` and `go test ./internal/connectors/...` timed out at 10 minutes with no failing package output before timeout.
