# TDD ledger — Workday REST parity wave05-r1

## Baseline

- Current bundle had 3 streams and 4 `api_surface.json` rows.
- Baseline conformance command passed: `go test ./internal/connectors/conformance -run 'TestConformance/workday-rest' -count=1`.
- The subissue-provided `go run ./cmd/connectorgen validate internal/connectors/defs/workday-rest` command walks nested `fixtures/` and `schemas/` as connector roots in this repo-local tool version; full defs validation with filtering was used for connector-specific evidence.

## Red checks observed

- First generated bundle failed conformance because `spec.json` used root `additionalProperties`, which the repo-local spec meta-schema rejects.
- First generated fixtures failed conformance because stream fixtures were not wrapped in the recorded `request`/`response` shape and the check fixture omitted `request.query.limit=1`.
- First generated write fixtures exposed path mismatches where body fields named `id` overwrote path-field samples; generation now preserves path-field values and skips duplicate body properties.

## Green checks after implementation

- `go run ./cmd/connectorgen validate internal/connectors/defs --json` filtered for `workday-rest`: 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/workday-rest' -count=1`: pass.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=10m`: pass after updating tracked golden transcripts for the new Workday provider command surface.
- `go vet ./internal/connectors/... ./internal/cli/...`, `go build ./cmd/pm`, `make connector-boundary`, `git diff --check`, and `GOFLAGS='-p=1' make verify`: pass.
