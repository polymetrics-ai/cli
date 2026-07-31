# TDD ledger — Zoom connector parity (#3110)

Manual GSD programming-loop fallback is active because `scripts/gsd prompt programming-loop ...` is not exposed by this repo-local adapter. The connector is data, not Go, so fail-first gates are bundle validation, fixture-backed conformance, focused CLI/help tests, generated-doc validation, and boundary checks.

## Red / baseline

Pre-change Zoom bundle was partial:

- `internal/connectors/defs/zoom/api_surface.json` listed 7 endpoint rows, not the 184 official operations in the landed audit/source.
- `metadata.json` had `capabilities.write=false` and no `writes.json`.
- Local connector snapshot from #3110: 3 streams, 3 stream fixtures, 0 write actions, 0 write fixtures, no certification contract, no generated CLI commands.
- A direct `go run ./cmd/connectorgen validate internal/connectors/defs/zoom` invocation is a known validator footgun: it treats `fixtures/` and `schemas/` as connector directories and fails on missing `metadata.json`; focused validation must run over `internal/connectors/defs` and inspect/filter the Zoom result.

Expected failing assertion before edits: exact operation parity cannot be true while Zoom has 3 implemented rows and 181 blocked/planned rows in the issue snapshot.

## Green targets

- `go run ./cmd/connectorgen validate internal/connectors/defs` reports 0 findings, with Zoom loading 184 `api_surface.json` rows, 48 streams, 91 write actions, generated CLI surface metadata, operations metadata, and certification metadata.
- `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1` passes for stream fixtures and write request shapes.
- Focused CLI tests (`go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`) pass after regenerated docs/help/goldens.
- `go build ./cmd/pm`, `make connector-boundary`, `make verify`, and `git diff --check` pass before commit.

## Evidence log

- `go run ./cmd/connectorgen validate internal/connectors/defs`: passed with 0 findings after Zoom compact regeneration.
- `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1`: passed.
- Focused CLI timeout diagnosis before scope correction: the initial bounded run did not hang in a single operation; it exhausted package wall-clock while recursive certification tests repeatedly loaded the full embedded connector catalog. Running `TestCertifyCLISingleConnectorTextMode` alone passed, proving the timeout was cumulative. Compacting Zoom generated JSON and limiting `cli_surface.json` to direct/binary commands reduced the overhead but `go test -timeout 20m ./internal/connectors/certify` still timed out near the end of the package.
- Scope correction requested after `426eb0e3b`: the shared `internal/connectors/bundleregistry/registry.go` cache is out of connector-local scope and was restored to the pre-commit parent contents (`git diff HEAD^ -- internal/connectors/bundleregistry/registry.go` is empty). No shared cache/runtime workaround remains in this branch.
- No-cache focused rerun passed: `go run ./cmd/connectorgen validate internal/connectors/defs` (0 findings), `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1`, and `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=30m` (413.4s).
- `make verify` without the cache change passed when run through the long-running shell surface (`EXIT=0`, including `go test -timeout 20m ./...`, docs validation, smoke, lint, connectorgen validation, boundary, and release workflow check). An earlier `ctx_execute` attempt was externally SIGTERM'd during the long test phase and left an orphaned local test process; after cleaning that process, the actual repository gate passed, so no genuine timeout dependency remains to carry in this connector-local branch.
- `git diff --check`: passed.
