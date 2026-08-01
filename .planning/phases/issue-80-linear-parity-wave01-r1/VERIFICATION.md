# Verification checklist — issue #80 Linear parity wave01-r1

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json` — passed, 0 findings. Note: direct `defs/linear` path is not a valid validator root because `schemas/` and `fixtures/` are interpreted as connector dirs.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/linear' -count=1` — passed.
- [x] `go test ./internal/connectors/engine ./internal/connectors/conformance -count=1` — passed.
- [x] `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` — updated generated CLI golden transcript after Linear became a connector command surface.
- [x] `go test ./...` — passed after golden transcript refresh.
- [x] `go vet ./...` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `make verify` — passed; re-run after generated Linear connector docs/catalog refresh also passed.
- [x] CLI help parity spot checks passed: `./pm help linear`, `./pm linear`, `./pm linear --help`, `./pm linear issues list --help`, `./pm linear issue add label --help`.
- [x] `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` then reverted unrelated generator drift and kept Linear docs/catalog deltas only.
- [x] `./pm docs validate --connectors-dir docs/connectors` — passed.
- [x] `git diff --check` — passed.
- [x] `git status --short` scoped to Linear connector artifacts, CLI help parity support/goldens, generated Linear connector docs/catalog, GSD evidence, and documentation housekeeping only; no out-of-scope shared runtime/provider files.
- [x] Confirm no live credentials/provider calls, no shared runtime edits, no merges.
- [x] Confirm issue addendum exists on #80 and #97-#103 without count changes.
- [x] Re-audit exact Linear operation totals from current official schema against `api_surface.json` after command metadata refresh — blob `e92dc40c31e3b6e3962f93fa1d8cbe91f3e83034` at commit `7ef4c5024f88667b2c85057ff4c905676c4a93c2` has 165 Query + 372 Mutation + 80 Subscription = 617 official root operations; 617 unique ledger rows; 0 missing; 0 extra.
- [x] Confirm every destructive/delete-like `writes.json` action has `confirm: "destructive"` and a canonical `cli_surface.json` command whose approval text names typed destructive confirmation — 122/122 writes have commands; 93/93 destructive actions have typed destructive command approval.
- [x] Re-run connector validation, Linear conformance, docs/golden checks, and broader local gates after parity-check fixes — passed `connectorgen validate`, Linear conformance, docs validate, CLI golden transcript test, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and `make verify` (with isolated GOCACHE to avoid shared-cache contention).
