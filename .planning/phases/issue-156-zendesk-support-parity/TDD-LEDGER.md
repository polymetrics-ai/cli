# TDD ledger — Issue 156 Zendesk Support parity

## Setup evidence

- Isolation verified with `pwd -P` and `git rev-parse --show-toplevel`: `/Users/karthiksivadas/.treehouse/cli-83d592/5/cli`.
- Branch created: `fm/cli-zendesk-support-parity-wave01-r1`.
- `no-mistakes doctor`: daemon running, repo initialized.
- `scripts/gsd doctor`: passed.
- Required `programming-loop` adapter command unavailable: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active.

## Red / baseline observations before production edits

- Official Zendesk Support OAS 2.0.0 inventory: 625 operations.
- Baseline bundle inventory: 33 streams, 27 write actions, no `operations.json`, no `cli_surface.json`.
- Baseline `api_surface.json`: 76 rows; 33 stream-covered, 27 write-covered, 16 legacy excluded; not in operation-ledger mode.
- Baseline comparison against official OAS: 587 official operations missing from local ledger and 37 local rows not present exactly in the Support OAS path set.

## Planned green checks

- `go run ./cmd/connectorgen validate internal/connectors/defs/zendesk-support`
- `go test ./internal/connectors/conformance -run 'TestConformance/zendesk-support' -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- `go build ./cmd/pm`
- `make connector-boundary`
- `git diff --check`

## Slice ledger

| Slice | Red evidence | Green evidence | Notes |
| --- | --- | --- | --- |
| Issue addendum | Parent/subissues lacked captain correction in their bodies. | `gh-axi` edit script updated #156-#163 idempotently with marker `zendesk-support-captain-policy-addendum-r1`. | Preserved count tables; no implemented-count edits. |
| Operation ledger | Baseline api surface omitted 587 official operations and used legacy `excluded` rows. | `api_surface.json` now uses `operation_ledger_version: 1`, inventories 625 official OAS operations exactly once, preserves 33 stream + 27 write covered surfaces, and adds 571 blocked official operation rows. `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 0 findings. | Six supplemental rows cover existing bundle streams absent from the Support OAS; they are not counted as official operations. |
| Operation/CLI metadata | Baseline had no `operations.json`/`cli_surface.json`. | Added 571 typed blocked operation contracts and 631 command metadata entries. `go test ./cmd/connectorgen -count=1` passed. | No executor claim; planned operation commands remain blocked by default. |
| Docs/conformance | Baseline docs mention only current 33/27 and legacy excluded categories. | Updated docs/metadata/generated Zendesk manual and skill. `go test ./internal/connectors/conformance -run 'TestConformance/zendesk-support' -count=1`, `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`, and docs validate passed. | Destructive/delete policy documented as in-scope with typed confirmation, not blanket unsafe exclusion. |
