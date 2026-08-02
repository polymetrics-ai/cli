# Verification checklist — Airtable official API parity

## Prior required gates

These gates passed before the review-hardening changes and are retained as historical evidence; the outer pipeline owns authoritative full-tree validation for the final tree.

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/airtable` — `connectorgen validate: 1 connector(s) checked, 0 findings` after marking the HyperDB `--primary-key` direct-read flag required to match the typed `body.primaryKeys` schema.
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

## Recovery reconciliation — 2026-08-02

- Forward-only recovery merge preserved local head `69f87c4976a0dbd225eb74d9768f4e74294e120e` and imported pipeline head `325d7ea1148969327eb369c0b3062edcf00e6cda` from the private no-mistakes recovery ref.
- Merge base `86d510927a05aa56b184bf5a8778b5444c69b9b1`; left-only inventory was the original Airtable parity commit, right-only inventory was the rebased Airtable parity commit, four Airtable no-mistakes review fixes, and current-main connector commits carried by the pipeline rebase.
- Changed-in-both conflicts were resolved to the pipeline Airtable/generated tree to preserve the review fixes, then amended only with `cli_surface.json` `required: true` on `hyperdb get-records --primary-key` because single-bundle validation proved the required typed body field was otherwise not reachable through CLI validation.
- Conflict proof before merge commit: no unmerged paths, `git diff --check`, `git diff --cached --check`, `go run ./cmd/connectorgen validate internal/connectors/defs/airtable`, `go test ./internal/connectors/conformance -run 'TestConformance/airtable' -count=1`, and `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` passed fixture-only.

## Results

Implemented fixture-only Airtable parity partition: 103 official OpenAPI operations tracked as 28 stream-backed GET/read/changefeed operations, 44 typed write actions, 1 HyperDB direct-read operation/CLI command, and 30 blocked operations. Comments use the exact official per-record endpoint, webhook replay uses the executable `limit=50`, and attachment upload remains blocked on `airtable-bounded-base64-upload-foundation` instead of exposing an unbounded write.

## Review-hardening gate

- [x] `go test ./internal/connectors/defs ./internal/connectors/conformance -run 'Airtable|Conformance/airtable' -count=1` — passed for both focused packages.
