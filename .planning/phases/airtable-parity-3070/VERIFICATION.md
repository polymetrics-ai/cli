# Verification checklist — Airtable official API parity

## Prior required gates

These gates passed before the review-hardening changes and are retained as historical evidence; the outer pipeline owns authoritative full-tree validation for the final tree.

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

Implemented fixture-only Airtable parity partition: 103 official OpenAPI operations tracked as 28 stream-backed GET/read/changefeed operations, 44 typed write actions, 1 HyperDB direct-read operation/CLI command, and 30 blocked operations. Comments use the exact official per-record endpoint, webhook replay uses the executable `limit=50`, and attachment upload remains blocked on `airtable-bounded-base64-upload-foundation` instead of exposing an unbounded write.

## Review-hardening gate

- [x] `go test ./internal/connectors/defs ./internal/connectors/conformance -run 'Airtable|Conformance/airtable' -count=1` — passed for both focused packages.
