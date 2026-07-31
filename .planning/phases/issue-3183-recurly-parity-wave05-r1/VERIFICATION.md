# Verification checklist — Recurly parity wave05 r1

## Final gate rerun after resume

- [x] Isolation confirmed: `pwd -P` and `git rev-parse --show-toplevel` both point at `/Users/karthiksivadas/.treehouse/cli-83d592/48/cli`; branch is `fm/cli-recurly-parity-wave05-r1`.
- [x] Inventory preserved at `/tmp/recurly-wave05-r1-snapshot-20260801023416` before final reruns.
- [x] Scope guard: modified/untracked files are limited to Recurly defs, Recurly generated connector docs, generated website connector catalog surfaces, CLI golden transcripts, and GSD planning artifacts. No shared runtime files are modified.
- [x] Official source count reproduced from Recurly OpenAPI v2021-02-25: 197 operations (GET 97, POST 42, PUT 35, DELETE 23).
- [x] Focused connectorgen validation for Recurly via single-connector temp defs root: `exit 0`, 0 findings, 0 warnings, 1 connector checked.
- [x] Full connectorgen validation: `go run ./cmd/connectorgen validate internal/connectors/defs --json` checked 549 connectors with 0 findings and 0 warnings.
- [x] Focused conformance: `go test ./internal/connectors/conformance -run 'TestConformance/recurly' -count=1` passed (`ok`, about 3.6s).
- [x] Focused CLI/dynamic/golden/docs tests passed: `go test -timeout 10m ./internal/cli -run '<focused dynamic/golden/docs regex>' -count=1` (`ok`, about 120s).
- [x] `go build ./cmd/pm` passed.
- [x] Connector docs validation: `./pm docs validate --connectors-dir docs/connectors` passed.
- [x] Boundary check: `make connector-boundary` passed with `outcome: clean`.
- [x] `git diff --check` passed after final edits.
- [x] Issue addendum marker previously verified exactly once on #3183-#3190.

## Schema / CLI correction

- [x] Direct-read CLI flags map only concrete required body leaves, not whole JSON objects or arrays.
- [x] `pm recurly gift cards preview --help` shows `--unit-amount (integer)` mapped to `body.unit_amount`, avoiding object/array passthrough and keeping direct read request construction typed.

## Make verify note

`make verify` was attempted before the final focused/full reruns. It repeatedly reached the `go test -timeout 20m ./...` target and timed out in unrelated timeout-heavy packages (`internal/cli` and/or `internal/connectors/certify`) while many other worktrees were concurrently running `make verify` / `go test` processes on this host. Current process evidence showed concurrent `/tmp/fm-cli-*` test binaries and other `make verify` / `go test -timeout 20m ./...` runs. This is recorded as an unrelated local resource-contention timeout, not a Recurly validation failure. The Recurly-specific and full connector validation/build/docs/boundary/diff gates above are green.

## Not run by design

- Live provider checks, credentialed Recurly calls, certification claims, pushes, PR updates, VPS, Thaalam changes, `/no-mistakes`, and shared-runtime edits were intentionally not run for this wave.
