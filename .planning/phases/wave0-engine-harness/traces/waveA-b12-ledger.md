# TDD ledger — T/B-12 certify report + cliharness

Package: `internal/connectors/certify` (disjoint from `internal/connectors/engine`; zero engine
imports per DECISIONS.md #2 float rule).

## T-12

Status: red-confirmed

Command:
```
go test ./internal/connectors/certify -run 'TestReport|TestHarness|TestScanForSecrets|TestMustKind|TestLoadReport' -v
```

Output:
```
polymetrics.ai/internal/connectors/certify: no non-test Go files in /Users/karthiksivadas/Development/polymetrics-cli-agents/connector-architecture-rewrite/internal/connectors/certify
FAIL	polymetrics.ai/internal/connectors/certify [build failed]
FAIL
```

Timestamp: 2026-07-02T07:53:10Z

Notes: test files authored first per PLAN.md T-12 —
`report_test.go` (marshal/unmarshal round-trip vs certification-design §A / DATA-MODEL §6 JSON
shape, `Save`/history append under `.polymetrics/certifications/history/<connector>/<timestamp>.json`,
`LoadReport` round-trip, missing-connector-name guard) and `cliharness_test.go` (in-process
`cli.Run(["init","--root",tmp,"--json"], ...)` drive via `certify.NewHarness`, envelope
kind+exit assertions, non-JSON run has nil envelope, unknown-command capture, typed
`*certify.KindMismatchError` on kind/exit mismatch, `ScanForSecrets` exact/base64/URL-encoded
detection, argv redaction of planted secret values in `CLIResult.ArgvRedacted`). No production
files exist yet, so the package fails to build — RED confirmed via compiler error rather than a
test assertion failure, which is the expected first-red shape for a from-scratch package.

## B-12

Status: green-confirmed

Implementation: `report.go` (Report/Capabilities/CapabilityResult/SyncModeResult/StageResult/
CLIStageInfo types + `Save`/`LoadReport`, history append keyed by
`StartedAt.UTC().Format("20060102T150405Z")`), `cliharness.go` (`Harness`/`NewHarness`/
`WithSecrets`, `Run` auto-injects `--root` when the caller omits it, parses the `--json` envelope
via `encoding/json`, computes `ArgvRedacted` by string-replacing every registered secret value
with `***`, `KindMismatchError` typed failure, `ScanForSecrets` checking exact / base64
(std + raw-std) / URL-encoded forms), `certify.go` (`Options`/`Runner`/`NewRunner`/`Run` skeleton
returning `ErrNotImplemented` — stage execution is T/B-14's `stages_source.go`, out of scope
here). Zero imports of `internal/connectors/engine` (grep-verified) and zero edits to
`internal/cli` — the harness adapts to `cli.Run`'s existing envelope/`--root`/`--json` behavior
without touching it.

Command:
```
go build ./... && go vet ./internal/connectors/certify && go test ./internal/connectors/certify -v
```

Output (tail):
```
=== RUN   TestReportSaveRequiresConnectorName
--- PASS: TestReportSaveRequiresConnectorName (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/certify	(cached)
```

18 tests, 18 PASS, 0 FAIL. `gofmt -l internal/connectors/certify` empty. `go vet ./...` (whole
repo) clean. `git status --porcelain` shows only `internal/connectors/certify/**` and this ledger
file as new paths under my ownership (plus an unrelated parallel Wave A agent's
`internal/connectors/engine/` and `traces/waveA-ledger.md`, untouched by this task).

Timestamp: 2026-07-02T07:56:00Z

