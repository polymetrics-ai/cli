# TDD Ledger: Gorgias Operation Ledger

Parent issue: #196  
Sub-issue: #200  
Branch: `feat/200-gorgias-operation-ledger`

## 2026-07-10 — plan checkpoint

- Task type: connector API surface operation-ledger metadata.
- Production behavior changed: no.
- GSD evidence:
  - `scripts/gsd doctor` passed.
  - `scripts/gsd prompt plan-phase "Issue #200 Gorgias operation ledger: account for all 114 official Gorgias operations as implemented stream/direct_read/write or blocked operation categories before production edits; preserve safety gates, no credentials, no raw generic write tools"` generated the planning workflow prompt.
  - `programming-loop` is not listed by `scripts/gsd list --json`; manual GSD universal runtime fallback remains active.
- Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

## Red evidence

Added `cmd/connectorgen/gorgias_api_surface_test.go` with focused Gorgias operation-ledger invariants, then ran:

```bash
go test ./cmd/connectorgen -run GorgiasAPISurfaceOperationLedger
```

Initial result: failed as expected with `operation_ledger_version = 0, want 1`. Current `internal/connectors/defs/gorgias/api_surface.json` has 11 legacy rows and no operation ledger, while the official captured baseline has 114 operations.

## Green evidence

Focused green commands after generating the 114-row ledger:

```bash
go test ./cmd/connectorgen -run GorgiasAPISurfaceOperationLedger
jq empty internal/connectors/defs/gorgias/api_surface.json .planning/phases/issue-197-gorgias-cli-surface-metadata/OFFICIAL-OPERATIONS.json
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/gorgias'
git diff --check
```

Results:

- Gorgias ledger metrics test passed.
- JSON parse check passed.
- Full connector definition validation passed: 547 connector(s) checked, 0 findings.
- Gorgias conformance test passed.
- Diff whitespace check passed.

## Refactor notes

- JSON-only ledger changes should not alter connector runtime behavior.
- Go test additions require `gofmt -w cmd internal` before final verification.
