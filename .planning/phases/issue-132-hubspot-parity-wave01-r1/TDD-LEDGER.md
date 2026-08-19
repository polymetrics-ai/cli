# TDD Ledger — HubSpot parity wave01 r1 (#132)

## Setup

- `scripts/gsd doctor`: pass.
- `scripts/gsd list`: pass.
- `scripts/gsd prompt plan-phase 132 --skip-research --tdd`: prompt rendered.
- `scripts/gsd prompt gsd-quick --full ...`: prompt rendered.
- Manual-GSD fallback: `programming-loop` is not registered in `scripts/gsd`; keep plan, TDD ledger, and verification checklist current manually.

## Red / green strategy

This slice is connector-local ledger/scaffold work, not runtime execution. Red evidence is validation absence/failure before the bundle exists; green evidence is schema/conformance validation after generating the bundle.

| Slice | Red evidence | Green target |
|---|---|---|
| Bundle absent | `go run ./cmd/connectorgen validate internal/connectors/defs/hubspot` fails because path/bundle is missing | `metadata.json`, `spec.json`, `streams.json`, `api_surface.json`, `operations.json`, `cli_surface.json`, and `docs.md` load and validate |
| Operation inventory | HubSpot has zero local operation rows despite parent count 3,118 | 3,118 unique `(method,path)` rows from the official spec collection commit are represented exactly once |
| Destructive policy | Prior issue text could be read as excluding DELETE/destructive operations | #132 and #134-#140 include the captain addendum; DELETE rows are modeled as in-scope blocked destructive operations with typed-confirmation notes |
| Shared foundations | Provider search/query and CDC foundations are open | Connector-local docs and operation reasons cite #2985/#2986/#2988 and make no implementation/certification claim |

## Evidence log

- Red captured: `go run ./cmd/connectorgen validate internal/connectors/defs/hubspot` exits 1 with `validate: read root: open .: no such file or directory` because the HubSpot bundle is absent.
- Green captured: generated connector-local HubSpot bundle with 3,118 unique official operation rows; `go run ./cmd/connectorgen validate internal/connectors/defs` reports 549 connectors checked and 0 findings; `go test ./internal/connectors/conformance -run 'TestConformance/hubspot' -count=1` passes.

## Refactor notes

- No shared runtime or generated hook/native files changed.
- `internal/cli/catalog_cli_test.go` and `internal/cli/testdata/golden_transcripts.json` were updated because adding a connector-owned CLI surface changes generated root connector command discovery and catalog count from 552 to 553.
- Connector docs/catalog artifacts were regenerated only for HubSpot and the connector catalog index.
- The generated ledger is intentionally metadata-only until future implementation lanes add executable streams, direct reads, and write actions with fixtures.
