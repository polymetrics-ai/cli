# TDD Ledger

Phase: wave0-engine-harness

Record failing test evidence before production code for every behavior-adding task.
Full RED/GREEN command output per task lives in the per-wave ledger files under `traces/`;
this index tracks status. Coordinator merges after each dispatch wave.

| Task | Status | Evidence | Tests green |
| --- | --- | --- | --- |
| T/B-01 schema validator | red-confirmed → green | traces/waveA-ledger.md §T-01 | 20 |
| T/B-02 interpolator + when | red-confirmed → green | traces/waveA-ledger.md §T-02 | 15 |
| T/B-03 bundle loader + defs | red-confirmed → green | traces/waveA-ledger.md §T-03 | 12 |
| T/B-04 typed errors + error_map | red-confirmed → green | traces/waveA-ledger.md §T-04 | 10 |
| T/B-12 certify report + cliharness | red-confirmed → green | traces/waveA-b12-ledger.md §T-12 | 18 |
| T/B-19 inventorygen + inventory.json | red-confirmed → green | traces/waveB-b19-ledger.md §T-19 | 19 |
| T/B-07 hook interfaces + registry | red-confirmed → green | traces/waveB-ledger.md §T-07 | 9 |
| T/B-05 auth selection | red-confirmed → green | traces/waveB-ledger.md §T-05 | 17 |
| T/B-06 paginators (6 types + SSRF guard) | red-confirmed → green | traces/waveB-ledger.md §T-06 | 21 |
| T/B-11 connectorgen validate\|gen\|new | red-confirmed → green | traces/waveC-b11-ledger.md §T-11 | 23 |
| T/B-08 read path | red-confirmed → green | traces/waveC-ledger.md §T-08 | 28 |
| T/B-09 write path | red-confirmed → green | traces/waveC-ledger.md §T-09 | 17 |
| T/B-10 connector assembly + Definition | red-confirmed → green | traces/waveD-b10-ledger.md §T-10 | pass |
| T/B-18 .golangci.yml + Makefile gates | red-artifact → green | traces/waveD-b18-ledger.md | make lint 0 issues |
| T/B-14 certify source stages vs sample | red-confirmed → green | traces/waveE-b14-ledger.md §T-14 | 21 |

Wave A gate (coordinator, 2026-07-02): `go build ./...` ok · engine+certify tests ok · gofmt clean ·
path guard clean (no tracked-file modifications outside plan). Noted deviations (agent-documented):
interpolate.go 331 lines vs ~150 target (CRLF guard + ResolveCheck + when-parser); `defs.go`
`//go:embed all:*` verified to tolerate the empty defs tree; engine coverage 80.0% at end of Wave A
(≥85% is a phase-exit gate, re-measured at V-21).
