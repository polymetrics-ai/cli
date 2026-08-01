# TDD Ledger — issue #3215 BambooHR parity

## Red / baseline

- `go run ./cmd/connectorgen validate`: pass (baseline whole-tree validation; single-dir invocation is not supported by this adapter and treats nested `schemas/`/`fixtures/` as connectors).
- `go test ./internal/connectors/conformance -run 'TestConformance/bamboo-hr' -count=1`: pass baseline.
- Current audit vs latest official OpenAPI before changes:
  - Official operations: 316 (310 path operations + 6 top-level OpenAPI webhooks).
  - Local `api_surface.json`: 340 rows.
  - Missing official rows: 23.
  - Stale local extra rows: 47.
  - Local implemented rows: 84 streams + 101 writes = 185.
  - Local blocked rows used legacy `excluded` rather than typed operation-ledger rows.
- Added red test: `go test ./cmd/connectorgen -run TestBambooHRSurfaceTracksCurrentOfficialInventory -count=1` failed before production edits on `operation_ledger_version = 0`.

## Green implementation

- Regenerated BambooHR connector-local operation ledger from the current official OpenAPI and top-level webhooks.
- Final count truth:
  - Official/local endpoint rows: 316.
  - Implemented rows: 296 = 138 streams + 9 direct reads + 149 writes.
  - Blocked/planned rows: 20 = 6 `binary_read`, 7 `admin_reverse_etl`, 1 `disallowed`, 6 `local_workflow`.
  - Legacy `excluded` rows: 0.
  - Stale local extras: 0.
  - Missing official rows: 0.
- Fixture/direct/write safety:
  - Streams and writes use synthetic non-secret fixtures only.
  - Direct-read POST root body schemas are closed.
  - Writes use closed record schemas and typed action names; destructive actions require explicit destructive confirmation.
  - Binary/download/export, multipart/file/form-login, and inbound webhook deliveries are blocked rather than exposed as generic escape hatches.

## Green verification

- `go test ./cmd/connectorgen -run TestBambooHRSurfaceTracksCurrentOfficialInventory -count=1`: pass.
- `go run ./cmd/connectorgen validate`: pass, `550 connector(s) checked, 0 findings`.
- `go test ./internal/connectors/conformance -run 'TestConformance/bamboo-hr' -count=1`: pass.
- `go test ./cmd/connectorgen -run 'BambooHR|CLISurface|APISurface|Connector|Gong|GitHub' -count=1`: pass.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=8m`: pass.
- `make verify`: pass on final pipefail run.

## Refactor / safety notes

- No live BambooHR calls were made and no credentials were requested, read, printed, summarized, or stored.
- No shared runtime behavior was changed.
- GSD `programming-loop` command was unavailable in this repo-local adapter; manual GSD/TDD loop was documented before production edits.

## no-mistakes review fix loop

- `no-mistakes status` found no active run owning `fm/cli-bamboohr-parity-wave05-r1`; `no-mistakes rerun` was attempted as requested and failed because the gate repository had no preserved branch ref for this branch.
- Started `no-mistakes axi run --intent ...`; review gated on three findings (destructive POST deletes, direct-read examples missing required flags, unsupported object/array-object CLI flags).
- `no-mistakes axi respond --action fix --findings F1,F2,F3 ...` was attempted; the pipeline fix agent failed with a WebSocket error and the run entered terminal `failed`, so follow-up fixes are made between runs and committed on top before rerunning no-mistakes.
