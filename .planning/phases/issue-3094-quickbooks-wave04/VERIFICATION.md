# Verification — QuickBooks Wave 04

## Required local gates

- [x] Focused QuickBooks connectorgen validation via temp defs root:
  `tmp=$(mktemp -d); cp -R internal/connectors/defs/quickbooks "$tmp/quickbooks"; go run ./cmd/connectorgen validate "$tmp"; rm -rf "$tmp"`
  - Result: `connectorgen validate: 1 connector(s) checked, 0 findings`
  - Note: `connectorgen validate` expects a definitions root containing connector subdirectories; the literal connector-dir argument treats `fixtures/` and `schemas/` as connector directories. The full definitions-root validation below also passed.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  - Result: `connectorgen validate: 549 connector(s) checked, 0 findings`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/quickbooks' -count=1`
  - Result: `ok   polymetrics.ai/internal/connectors/conformance`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=12m`
  - Result: `ok   polymetrics.ai/internal/cli   143.513s`
- [x] `go build ./cmd/pm`
  - Result: passed with no output.
- [x] `make connector-boundary`
  - Result: outcome `clean`, 0 findings, 0 warnings.
- [x] `GOFLAGS='-p=1' make verify`
  - Result: passed; final output ended with `homebrew release notification assertions passed`.
  - Note: serial package execution avoids the local process-killer/SIGTERM seen with parallel `go test ./...` in this constrained worktree.
- [x] `git diff --check`
  - Result: passed with no output.

## Focused gates

- [x] `go test ./cmd/connectorgen -run TestQuickBooksAPISurfaceOperationLedgerMetrics -count=1`
  - Result: `ok   polymetrics.ai/cmd/connectorgen`
- [x] `go test ./internal/connectors/hooks/quickbooks -count=1`
  - Result: `ok   polymetrics.ai/internal/connectors/hooks/quickbooks`

## Safety assertions

- [x] No live provider calls.
- [x] No credential requests or secret literals.
- [x] No generic raw HTTP/query/file/shell escape hatches.
- [x] Reverse-ETL operations remain blocked unless typed schemas/fixtures/approval evidence exist.
- [x] Fixture-only certification status remains uncertified.

## Results

Local gates passed. A reviewer subagent re-review reported no blocker findings after fixes for realm ID handling, token refresh auth shape, safety-cap docs, CLI flags, and upload-row classification. No push, PR, no-mistakes, live-provider call, credential collection, or remote issue comment was performed.
