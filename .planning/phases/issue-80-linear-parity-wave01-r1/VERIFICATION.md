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
- [x] `git status --short` scoped to Linear connector artifacts, generated CLI golden transcript, generated Linear connector docs/catalog, and GSD evidence only (149 status entries; no out-of-scope shared runtime/provider files).
- [x] Confirm no live credentials/provider calls, no shared runtime edits, no merges.
- [x] Confirm issue addendum exists on #80 and #97-#103 without count changes.
