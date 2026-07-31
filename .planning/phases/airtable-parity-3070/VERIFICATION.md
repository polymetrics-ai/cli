# Verification checklist — Airtable official API parity

## Required gates

- [!] `go run ./cmd/connectorgen validate internal/connectors/defs/airtable` — current `connectorgen validate` treats the argument as a directory of bundle directories, so this exact single-bundle path validates `fixtures/` and `schemas/` as fake connectors and exits 1 with missing `metadata.json`. No tooling/runtime behavior was changed for this local-only connector wave.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — `connectorgen validate: 549 connector(s) checked, 0 findings`.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/airtable' -count=1` — `ok polymetrics.ai/internal/connectors/conformance`.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — `ok polymetrics.ai/internal/cli`.
- [x] `go build ./cmd/pm`.
- [x] `make connector-boundary` — outcome `clean`.
- [x] `make verify` — passed on cached retry after one transient full-suite timeout in `internal/connectors/certify` under parallel `go test -timeout 20m ./...`; package passed standalone and final `make verify` completed successfully.
- [x] `git diff --check`.

## Fixture-only safety

- No live provider calls.
- No credentials requested or used.
- No Airtable writes executed.
- Certification metadata is fixture/candidate-only; no live certification claim was made.
- No push, PR, `/no-mistakes`, VPS, Thaalam, or provider-side operations were run.

## Results

Implemented fixture-only Airtable parity partition: 103 official OpenAPI operations tracked as 31 stream-backed GET/read/changefeed operations, 70 typed write actions (including attachment upload), 1 HyperDB direct-read operation/CLI command, and 1 blocked Sync API CSV import with an explicit typed-runtime reason.
