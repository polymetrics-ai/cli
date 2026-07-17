# Phase 423 TDD Ledger

Issue: #423 — nativize perf namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-code-style`.

Repo skill gap: `.pi/skills/go-implementation/SKILL.md` was required by worker instructions but is absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors:

- `golang-how-to`: CLI command tree routes to `golang-spf13-cobra` + `golang-cli`; tests route to `golang-testing`; args/I/O route to `golang-security` + `golang-safety`.
- `golang-cli`: preserve exit codes, stdout/stderr discipline, CLI unit tests, and machine-readable output.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior/public contract over implementation-only details.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context when propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording; application CLI help is primary documentation.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags guidance for `StringArray`, `NoOptDefVal`, and unknown-flag compatibility.
- `golang-security`: trust-boundary questions #1-#3; no secrets; command args are untrusted; no runtime services started for tests.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.
- `golang-code-style`: early returns, clear small helpers, semantic line breaks.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 423 --skip-research >/tmp/gsd-plan-phase-423.prompt
scripts/gsd prompt programming-loop init --phase 423 --dry-run >/tmp/gsd-programming-loop-423.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-423.prompt` (10664 bytes).
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md` + universal runtime loop.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1` | Fail | Native-subtree tests fail because `perf` remains legacy; behavior tests already preserve current flag semantics. || 2 | Green | `gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/perf_cli_test.go`; `go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1` | Pass | Native perf parser green; flag semantics, runtime config use, bare help, and invalid action preserved. |
| 3 | Refactor | `gofmt -w cmd internal`; `go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1`; `go test ./internal/cli/ -run Certify -count=1`; `go vet ./...`; `go build ./cmd/pm` | Pass | Golden-focused gate, certify re-entrancy smoke, vet, and build preserved. |
| 4 | Full gate | `go test ./...`; `make verify`; `go build ./cmd/pm`; runtime help/docs/website/diff checks | Pass | Full local gates, CLI parity checks, and diff guards passed; no go.mod/go.sum diff. |

## Planned red tests

- `TestPerfCommandIsNativeCobraSubtree`: current wrapper should fail because `perf` remains legacy; expected native `compare`/`sync-modes` subcommands, declared `StringArray` flags, `NoOptDefVal="true"`, unknown-flag whitelist, and no-file completion seams are missing.
- `TestPerfCompareFlagFormsPreserveLegacySemantics`: current metadata path should fail until pflag declarations and normalization exist; behavior cases cover space/equals forms, repeated scalar last-wins, bare bool/value sentinels, unknown flags, extra args, late globals, JSON envelope preservation, and runtime config endpoint use.
- `TestPerfSyncModesFlagFormsPreserveLegacySemantics`: records space/equals/repeated/bare-value semantics and output envelope preservation.
- `TestPerfBareAndInvalidActionSemantics`: bare namespace help must exit 0 and invalid action must remain usage exit 2.

## Exact red outputs

```bash
go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1
```

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestPerfCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:181: perf command must use native Cobra flag parsing
redis: 2026/07/18 00:29:07 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:08 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:08 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:08 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:10 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:10 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:11 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
redis: 2026/07/18 00:29:11 pool.go:724: redis: connection pool: failed to dial after 5 attempts: dial tcp 127.0.0.1:2: connect: connection refused
FAIL
FAIL	polymetrics.ai/internal/cli	12.967s
FAIL
```

Red note: loopback runtime-check connection-refused messages are expected from `--runtime` flag validation against local test endpoints; no services were started and no secrets are involved.

## Review-fix cycle plan — 2026-07-18

Additional skills loaded for review-fix: `golang-lint`, `golang-benchmark`, `golang-performance`. Existing loaded skills still apply: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-code-style`, `gsd-core`, `caveman`. Repo skill gap remains: `.pi/skills/go-implementation/SKILL.md` missing (`ENOENT`).

GSD review-fix command evidence:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 423 --skip-research >/tmp/gsd-plan-phase-423-review-fix.prompt
scripts/gsd prompt programming-loop init --phase 423 --dry-run >/tmp/gsd-programming-loop-423-review-fix.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-423-review-fix.prompt` (10664 bytes).
- `programming-loop`: still blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback remains active.

Planned red tests before production edits:

- `TestPerfCompareRejectsInvalidAndOversizedIterations`: invalid and oversized `--iterations` must exit 3 with JSON `validation_error`; default and small values remain green.
- `TestPerfSyncModesRejectsInvalidAndOversizedRecords`: invalid and oversized `--records` must exit 3 with JSON `validation_error`; default and small values remain green.
- `TestPerfManualMentionsValidationAndRuntimeMetadata`: embedded perf manual must list exit 3 and describe `runtime_report` metadata/redaction/topology.
- `internal/perf` tests for direct API max guards and runtime-backed error redaction via the shared registry.
- `internal/runtimecheck` test capturing process stderr around a failing Dragonfly check; raw `redis:` diagnostics must not bypass the logger.

Expected first red command set:

```bash
go test ./internal/cli/ -run 'Perf' -count=1
go test ./internal/perf -count=1
go test ./internal/runtimecheck -count=1
```

## Review-fix exact red output

```bash
go test ./internal/cli/ -run 'Perf' -count=1
go test ./internal/perf -count=1
go test ./internal/runtimecheck -count=1
```

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/perf_cli_test.go:134:48: undefined: perf.MaxCompareIterations
internal/cli/perf_cli_test.go:161:48: undefined: perf.MaxSyncModeRecords
FAIL	polymetrics.ai/internal/cli [build failed]
FAIL
# polymetrics.ai/internal/perf [polymetrics.ai/internal/perf.test]
internal/perf/perf_test.go:30:69: undefined: MaxCompareIterations
internal/perf/perf_test.go:40:85: undefined: MaxSyncModeRecords
internal/perf/perf_test.go:50:23: undefined: dependencyFreeRunner
internal/perf/perf_test.go:51:22: undefined: runtimeBackedRunner
internal/perf/perf_test.go:52:22: undefined: runtimeDoctor
internal/perf/perf_test.go:54:3: undefined: dependencyFreeRunner
internal/perf/perf_test.go:55:3: undefined: runtimeBackedRunner
internal/perf/perf_test.go:56:3: undefined: runtimeDoctor
internal/perf/perf_test.go:64:2: undefined: dependencyFreeRunner
internal/perf/perf_test.go:67:2: undefined: runtimeDoctor
internal/perf/perf_test.go:67:2: too many errors
FAIL	polymetrics.ai/internal/perf [build failed]
FAIL
ok  	polymetrics.ai/internal/runtimecheck	0.538s
```

Red note: runtimecheck stderr-capture hardening test already passed with the short timeout; the accepted diagnostic-routing fix remains in scope because prior perf red output showed raw go-redis stderr diagnostics during runtime checks. No secrets were involved.

## Review-fix exact green outputs

```bash
go test ./internal/cli/ -run 'Perf' -count=1
go test ./internal/perf -count=1
go test ./internal/runtimecheck -count=1
```

```text
ok  	polymetrics.ai/internal/cli	11.366s
ok  	polymetrics.ai/internal/perf	1.061s
ok  	polymetrics.ai/internal/runtimecheck	0.426s
```

```bash
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir "$TMP_CONNECTORS"
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli/ -run TestGoldenTranscripts -count=1
npm --prefix website run gen:docs
go test ./internal/cli/... -run 'Perf|GoldenDocs|GoldenTranscripts' -count=1
```

```text
Generated docs in docs/cli and connector docs in /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.eG67n91Itt
ok  	polymetrics.ai/internal/cli	10.090s
Wrote 11 docs pages to lib/docs.generated.ts.
ok  	polymetrics.ai/internal/cli	20.859s
```

```bash
go test ./internal/cli/...
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
go test ./internal/perf -count=1
go test ./internal/runtimecheck -count=1
go test ./...
make verify
git diff -- go.mod go.sum
git diff --check origin/feat/cli-architecture-v2...HEAD
```

```text
ok  	polymetrics.ai/internal/cli	186.954s
# gofmt, go vet, go build emitted no output and exited 0
ok  	polymetrics.ai/internal/perf	1.104s
ok  	polymetrics.ai/internal/runtimecheck	0.336s
go test ./... pass; slow packages included internal/cli 190.154s, internal/connectors/certify 343.247s, internal/perf 1.644s, internal/runtimecheck 1.440s
make verify pass; tail included:
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
# go.mod/go.sum diff and diff --check emitted no output
```

```bash
./pm help perf
./pm perf
./pm perf --help
./pm perf --json
./pm perf compare --iterations nope --json
./pm perf compare --iterations 1001 --json
./pm perf sync-modes --records nope --json
./pm perf sync-modes --records 100001 --json
./pm perf compare --iterations 1 --runtime=false --json
./pm perf sync-modes --records 5 --json
./pm --root "$root" --json perf compare --iterations 1 --runtime
```

```text
help/bare/--help stdout_bytes=1555 stderr_bytes=0; JSON manual stdout_bytes=1695 stderr_bytes=0
invalid/oversized iterations and records exited 3 with kind=Error category=validation code=validation_error
compare kind=PerformanceComparison iterations=1 records=3 runtime_report=False
sync kind=SyncModeBenchmark records=5 result_records=[5, 5, 5, 5, 5]
runtime compare exit=0; runtime_report endpoints postgres='postgres://127.0.0.1:1/polymetrics' dragonfly='127.0.0.1:2' temporal='127.0.0.1:3'; statuses={'postgres': 'error', 'dragonfly': 'error', 'temporal': 'error'}; stderr_bytes=0 raw_redis=0
```

## Exact green outputs

```bash
gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/perf_cli_test.go
go test ./internal/cli/ -run 'Perf|CobraRouterShell' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	13.101s
```

```bash
gofmt -w cmd internal
go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1
go vet ./...
go build ./cmd/pm
```

```text
ok  	polymetrics.ai/internal/cli	18.543s
# go vet and go build emitted no output and exited 0
```

```bash
go test ./internal/cli/ -run Certify -count=1
```

```text
ok  	polymetrics.ai/internal/cli	91.433s
```

```bash
go test ./...
```

```text
ok  	polymetrics.ai/cmd/connectorgen	6.152s
ok  	polymetrics.ai/cmd/iconregistrygen	0.726s
ok  	polymetrics.ai/cmd/pm	0.846s
ok  	polymetrics.ai/cmd/prissueguard	0.332s
ok  	polymetrics.ai/internal/agentmode	1.160s
ok  	polymetrics.ai/internal/app	20.112s
ok  	polymetrics.ai/internal/cli	179.296s
ok  	polymetrics.ai/internal/connectors/certify	367.560s
# pass; remaining packages also passed in terminal run.
```

```bash
make verify
```

```text
pass; completed gofmt, tidy-check, vet, go test -timeout 20m ./..., go build ./cmd/pm,
docs validate, local smoke flow, golangci-lint, and connectorgen validate.
Terminal tail:
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
```

```bash
go build ./cmd/pm
```

```text
# no output; command exited 0
```

```bash
./pm help perf
./pm perf
./pm perf --help
./pm perf --json
./pm perf bogus --json
./pm perf compare --iterations 1 --runtime=false --json
./pm perf sync-modes --records 5 --json
```

```text
help/bare/--help byte-identical; help bytes=874; JSON manual bytes=1004;
invalid action exit=2; stderr=error: unknown command "bogus" for "pm perf";
compare kind=PerformanceComparison iterations=1 records=3 runtime_backed absent;
sync-modes kind=SyncModeBenchmark records=5 and all result records=5.
```

```bash
./pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"
diff -ru docs/cli "$TMP_DOCS/cli"
./pm docs validate --connectors-dir docs/connectors
npm --prefix website run gen:docs
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

```text
Generated docs in <tmp>/cli and connector docs in <tmp>/connectors
# diff -ru: no output
Validated connector docs in docs/connectors
Wrote 11 docs pages to lib/docs.generated.ts.
# git diff checks: no output; go.mod/go.sum diff empty
```
