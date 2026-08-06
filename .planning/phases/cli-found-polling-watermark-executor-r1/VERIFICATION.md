# VERIFICATION — polling-watermark changefeed executor

## Planned evidence

| Area | Command/check | Status |
| --- | --- | --- |
| Red/green ledger | focused test output before/after production edits | passed — `TDD-LEDGER.md` |
| Connector contracts | `go test ./internal/connectors -count=1` | passed |
| Engine executor and bundle tests | `go test ./internal/connectors/engine -count=1` | passed |
| CLI regression package | `go test ./internal/cli -count=1` | pending |
| Formatting | `gofmt -w` changed Go files; `gofmt -l cmd internal` | pending |
| Vet/build | `go vet ./...`; `go build ./cmd/pm` | pending |
| Repository gates | individual `make verify` gates from `AGENTS.md` | pending |
| Capability preflight | real implemented-command/preflight sweep via changed package tests | pending |
| GSD review | inline review recorded after implementation | pending |

`go test ./...` and `make verify` will not run as single commands in this
per-command environment; CI carries the complete repository suite.

## Explicitly not applicable unless implementation proves otherwise

- Runtime help, bare namespace behavior, command help, CLI manual, website,
  generated manual, and completions: no command/flag/output text is planned.
- Credentialed checks, live provider/database calls, runtime services, reverse
  ETL execution, and wall-clock sleeps.
- Production connector declarations: task constraints require test-only
  bundles.
