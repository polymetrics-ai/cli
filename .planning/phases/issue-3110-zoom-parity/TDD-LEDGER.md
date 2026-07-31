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
- Focused CLI timeout diagnosis: the initial bounded run did not hang in a single operation; it exhausted package wall-clock while recursive certification tests repeatedly loaded the full embedded connector catalog. Running `TestCertifyCLISingleConnectorTextMode` alone passed, proving the timeout was cumulative. Compacting Zoom generated JSON and limiting `cli_surface.json` to direct/binary commands reduced the overhead but `go test -timeout 20m ./internal/connectors/certify` still timed out near the end of the package.
- Green timeout fix: cached immutable embedded bundle loading in `internal/connectors/bundleregistry.New()` while still constructing a fresh registry per call. `TestCertifyCLISingleConnectorTextMode` dropped from package-timeout risk to ~6s, and `go test -timeout 20m ./internal/connectors/certify -count=1 -v` passed in 35.4s.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=30m`: passed in 24.1s.
- `go build ./cmd/pm`: passed.
- `make connector-boundary`: passed, outcome clean with existing documented exceptions only.
- `make verify`: passed through release-workflow-check; no failure markers in `.tmp/zoom-make-verify.log`.
- `git diff --check`: passed.
- `gh-axi` captain-policy addendum marker posted and verified present on #3110-#3117.
