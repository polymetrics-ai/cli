# TDD Ledger: Gorgias Stream Runner

Parent issue: #196  
Sub-issue: #199  
Branch: `feat/199-gorgias-stream-runner`

## 2026-07-10 — plan checkpoint

- Task type: connector declarative stream expansion.
- Production behavior changed: no.
- GSD evidence:
  - `scripts/gsd prompt plan-phase "Issue #199 Gorgias stream runner: implement safe ETL stream coverage for list/read-sweep endpoints after #200 operation ledger; no credentials, no writes, no raw tools"` generated the planning workflow prompt.
  - Manual GSD universal runtime fallback remains active because `programming-loop` is unavailable in this repo-local command set.
- Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

## Red evidence

Pending: add focused Gorgias stream-runner regression test, then run:

```bash
go test ./cmd/connectorgen -run 'Gorgias(APISurfaceOperationLedger|StreamRunner)'
```

Expected initial failure: current parent baseline exposes 4 Gorgias streams and 4 `covered_by.stream` API rows; #199 target is 24 safe stream rows.

## Green evidence

Pending implementation.

## Refactor notes

- Declarative JSON/schema/fixture changes should not alter write behavior.
- Go test additions require `gofmt -w cmd internal` before final verification.
