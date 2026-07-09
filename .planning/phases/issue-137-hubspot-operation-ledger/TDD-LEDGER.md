# TDD Ledger: Issue #137 HubSpot Operation Ledger

Date: 2026-07-09

## GSD/TDD mode

- Desired command: `scripts/gsd prompt programming-loop init --phase issue-137-hubspot-operation-ledger --dry-run`.
- Result: blocked because the pinned registry does not contain `programming-loop`.
- Active fallback: manual universal programming loop with `scripts/gsd prompt plan-phase issue-137-hubspot-operation-ledger --skip-research` evidence.

## Red-test plan

Before production edits:

1. Add a failing HubSpot ledger test that expects:
   - `operation_ledger_version: 1`.
   - 3,060 unique official method/path operations.
   - method counts GET 1,038 / POST 1,314 / PUT 169 / PATCH 232 / DELETE 307.
   - no duplicate method/path rows.
   - no legacy `excluded` rows.
2. Add a failing validator/schema test for the app-operation ledger models needed by HubSpot classification (`stream_etl`, `query_etl`, `reverse_etl`, `binary_write` if needed) before adding them to the schema/validator vocabulary.

## Red evidence

- `go test ./cmd/connectorgen -run 'HubSpotOperationLedgerOfficialCounts|APISurfaceOperationLedgerAppCandidateModels' -count=1` failed before production edits:
  - HubSpot endpoint count was `0`, expected `3060`.
  - `stream_etl` app-candidate operation model was rejected by `api_surface.schema.json` enum.

## Green evidence

Passed after implementation:

- `gofmt -w cmd/connectorgen/main_test.go cmd/connectorgen/validate.go`
- `go test ./cmd/connectorgen -run 'HubSpot|APISurfaceOperationLedger' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 548 connector(s), 0 findings.
- `python3 -m json.tool internal/connectors/defs/hubspot/api_surface.json >/dev/null`
- `go run ./cmd/pm help connectors`
- `go run ./cmd/pm connectors inspect hubspot --json`
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- Full gate command passed: `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify && go run ./cmd/connectorgen validate internal/connectors/defs`.

## Refactor evidence

- Expanded operation-ledger blocked-candidate vocabulary to include `stream_etl`, `query_etl`, `reverse_etl`, and `binary_write`, preserving blocked-by-default semantics in both connectorgen and static conformance checks.
- Generated deterministic HubSpot `api_surface.json` from the official public OpenAPI collection: 401 OpenAPI files, 4,396 raw operations, 3,060 unique method/path operations.
- Classification counts: `stream_etl` 244, `query_etl` 223, `direct_read` 759, `reverse_etl` 850, `binary_read` 30, `binary_write` 31, `sensitive_reverse_etl` 40, `admin_reverse_etl` 291, `destructive_action` 556, `deprecated` 22, `disallowed` 14.

## Safety/TDD notes

- No live credentials.
- No live HubSpot API calls.
- Temporary official spec clone/read is public and credential-free.
- Do not create executable generic write actions for mutation endpoints in this ledger-only slice.
- Every mutation classification must remain blocked by default until a named reverse ETL action with fixed schema and plan → preview → approval → execute exists.
- Every binary classification must remain blocked by default until bounded destination/max-bytes policy exists.
