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

Added `TestGorgiasStreamRunnerReadSweep` in `cmd/connectorgen/gorgias_api_surface_test.go`, then ran:

```bash
go test ./cmd/connectorgen -run 'Gorgias(APISurfaceOperationLedger|StreamRunner)'
```

Initial result: failed as expected because the current parent baseline exposes only `customers`, `messages`, `satisfaction_surveys`, and `tickets`; #199 target is 24 safe stream rows plus matching `covered_by.stream` API rows.

## Green evidence

Focused green commands after adding the 24-stream read sweep:

```bash
go test ./cmd/connectorgen -run 'Gorgias(APISurfaceOperationLedger|StreamRunner)'
jq empty internal/connectors/defs/gorgias/streams.json internal/connectors/defs/gorgias/api_surface.json internal/connectors/defs/gorgias/cli_surface.json internal/connectors/defs/gorgias/schemas/*.json
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/gorgias'
git diff --check
```

Results:

- Gorgias API ledger and stream-runner tests passed.
- JSON parse checks passed.
- Full connector definition validation passed: 547 connector(s) checked, 0 findings.
- Gorgias conformance test passed.
- Diff whitespace check passed.

## Broad verification evidence

Final broad gate command:

```bash
gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify
```

Result: passed. `make verify` included docs validation, smoke test, golangci-lint, and `go run ./cmd/connectorgen validate internal/connectors/defs` with 547 connector(s), 0 findings.

## Refactor notes

- Declarative JSON/schema/fixture changes did not alter write behavior.
- `gofmt -w cmd internal` ran after Go test edits; no tracked changes remained after final broad verification.
