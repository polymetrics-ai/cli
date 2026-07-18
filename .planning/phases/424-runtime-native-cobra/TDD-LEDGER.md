# Phase 424 TDD Ledger

Issue: #424 — nativize runtime namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-lint`, `golang-code-style`.

Repo skill gap: `.pi/skills/go-implementation/SKILL.md` was required by worker instructions but is absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors:

- `golang-how-to`: CLI command tree routes to `golang-spf13-cobra` + `golang-cli`; tests route to `golang-testing`; args/runtime I/O route to `golang-security`, `golang-safety`, `golang-context`, and `golang-concurrency` as applicable.
- `golang-cli`: preserve exit codes, stdout/stderr discipline, CLI unit tests, and machine-readable output.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior/public contract over implementation-only details.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context when propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording; application CLI help is primary documentation.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags guidance for `StringArray`, `NoOptDefVal`, and unknown-flag compatibility.
- `golang-security`: trust-boundary questions #1-#3; no secrets; command args are untrusted; no runtime services started for tests.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.
- `golang-context`: #1 propagate same context, #3 never store context in structs, #5 cancel ownership when creating contexts.
- `golang-concurrency`: #1 goroutines need exits and #7 select includes `ctx.Done()` when adding concurrent work; no new goroutines planned.
- `golang-database`: runtime PostgreSQL docs/checks use parameterized/runtime-gated defaults; no schema or live credential changes.
- `golang-lint`: run `go vet ./...` and avoid suppressions/quality-gate reduction.
- `golang-code-style`: early returns, clear small helpers, semantic line breaks.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 424-runtime-native-cobra --skip-research >/tmp/gsd-plan-phase-424-runtime-native-cobra.prompt
scripts/gsd prompt programming-loop init --phase 424-runtime-native-cobra --dry-run >/tmp/gsd-programming-loop-424-runtime-native-cobra.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-424-runtime-native-cobra.prompt` (10739 bytes).
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md` + universal runtime loop.

Review-fix command refresh:

```bash
scripts/gsd doctor && scripts/gsd list
scripts/gsd prompt plan-phase 424-runtime-native-cobra --skip-research >/tmp/gsd-plan-phase-424-runtime-native-cobra-review-fix.prompt
scripts/gsd prompt programming-loop init --phase 424-runtime-native-cobra --dry-run >/tmp/gsd-programming-loop-424-runtime-native-cobra-review-fix.prompt
```

Result:

- `doctor && list`: pass; registry listed 69 commands.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-424-runtime-native-cobra-review-fix.prompt` (10739 bytes).
- `programming-loop`: still blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback continues.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1` | Fail | `runtime` remains a legacy wrapper and native `doctor` subcommand is missing. |
| 2 | Green | `gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/runtime_cli_test.go`; `go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1` | Pass | Native runtime parser green; bare help, JSON manual, invalid action category, unknown flags, extra args, config endpoints, and redaction preserved. |
| 3 | Refactor | `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`; `go test ./internal/cli/...`; `go vet ./...`; `go build ./cmd/pm` | Pass | Focused/golden, full internal CLI package, vet, and build green. |
| 4 | Full gate | `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`; runtime help/docs/website/diff checks | Pass | Full local gates, CLI parity checks, docs generator diff, website docs generator, and go.mod/go.sum diff guard passed. |
| 5 | Review-fix planning | Update PLAN/TDD-LEDGER/VERIFICATION/RUN-STATE before code | Green | PR #460 review-fix plan recorded; no production code touched. |
| 6 | Review-fix red | `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`; `go test ./internal/runtimecheck/... -count=1` | Fail | Missing-value pflag errors mapped to internal exit 1; DragonflyDB/Temporal endpoints reported raw userinfo/query/fragment. |
| 7 | Review-fix green | `gofmt -w ...`; focused CLI/runtimecheck tests; golden refresh; website docs gen | Pass | Cobra parse errors now usage; endpoint sanitizer green; runtime help/docs/website/goldens intentionally updated. |
| 8 | Review-fix full gate | `gofmt -w cmd internal`; `go vet ./...`; `go test -timeout 20m ./...`; `go build ./cmd/pm`; `make verify`; diff/docs/runtime CLI checks | Pass | Full local gates, docs-generate diff, website generator, malformed flag smoke, and endpoint sanitizer smoke passed. |

## Planned red tests

- `TestRuntimeCommandIsNativeCobraSubtree`: current wrapper should fail because `runtime` remains legacy; expected native `doctor` subcommand, unknown-flag whitelist, and no-file completion seam are missing.
- `TestRuntimeDoctorNativeCobraPreservesLegacySemantics`: behavior cases cover doctor JSON, unknown flag tolerance, extra args tolerance, late global `--json`, late global `--root`, config-file endpoints, and no raw secret leakage.
- `TestRuntimeBareHelpAndInvalidActionSemantics`: bare namespace help must exit 0, `--help` must render canonical docs, JSON manual must emit `CommandManual`, and invalid action must remain usage exit 2.

## Review-fix planned red tests

- `TestRuntimeMalformedKnownGlobalFlagsAreUsageErrors`: `--json runtime --root` and `--json runtime doctor --root` exit 2 with JSON category `usage`, not internal exit 1.
- Extend `TestCobraRouterShellMapsGenuineCobraParseErrorsToUsage`: genuine Cobra/pflag parse errors (unknown flag, missing required flag value, invalid bool value) classify as usage while `cobraLegacyError` remains internal when a legacy handler returns matching text.
- `TestRedactedConfigSanitizesAllRuntimeEndpoints`: `RedactedConfig` strips PostgreSQL userinfo/query/fragment and also strips DragonflyDB/Temporal userinfo/query/fragment/control chars.
- `TestRuntimeCheckResultsSanitizeServiceEndpointsAndErrors`: DragonflyDB/Temporal `CheckResult.Endpoint` and `Error` fields do not leak userinfo/query/fragment/control chars from configured endpoints.

## Exact red outputs

```bash
go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1
```

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestRuntimeCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:181: runtime command must use native Cobra flag parsing
FAIL
FAIL	polymetrics.ai/internal/cli	11.563s
FAIL
```

## Review-fix exact red outputs

```bash
go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1
```

```text
--- FAIL: TestCobraRouterShellMapsGenuineCobraParseErrorsToUsage (0.00s)
    --- FAIL: TestCobraRouterShellMapsGenuineCobraParseErrorsToUsage/missing_known_flag_value (0.00s)
        cobra_router_test.go:465: category = internal, want usage for genuine Cobra parse error "flag needs an argument: --root"
    --- FAIL: TestCobraRouterShellMapsGenuineCobraParseErrorsToUsage/invalid_bool_value (0.00s)
        cobra_router_test.go:465: category = internal, want usage for genuine Cobra parse error "invalid argument \"maybe\" for \"--json\" flag: strconv.ParseBool: parsing \"maybe\": invalid syntax"
--- FAIL: TestRuntimeMalformedKnownGlobalFlagsAreUsageErrors (0.00s)
    --- FAIL: TestRuntimeMalformedKnownGlobalFlagsAreUsageErrors/runtime_missing_root_value (0.00s)
        runtime_cli_test.go:24: Run([--json runtime --root]) exit = 1, want 2; stdout={..."category": "internal"..."message": "flag needs an argument: --root"...}
    --- FAIL: TestRuntimeMalformedKnownGlobalFlagsAreUsageErrors/runtime_doctor_missing_root_value (0.00s)
        runtime_cli_test.go:24: Run([--json runtime doctor --root]) exit = 1, want 2; stdout={..."category": "internal"..."message": "flag needs an argument: --root"...}
FAIL
FAIL	polymetrics.ai/internal/cli	17.501s
FAIL
```

```bash
go test ./internal/runtimecheck/... -count=1
```

```text
--- FAIL: TestRedactedConfigSanitizesAllRuntimeEndpoints (0.00s)
    runtimecheck_test.go:43: dragonfly_addr leaked "pm-review-fix-endpoint-secret" in "redis://user:pm-review-fix-endpoint-secret@dragonfly.local:6379/0?token=pm-review-fix-endpoint-secret#fragpm-review-fix-endpoint-secret"
--- FAIL: TestRuntimeCheckResultsSanitizeServiceEndpointsAndErrors (0.02s)
    runtimecheck_test.go:67: dragonfly endpoint = "redis://user:pm-review-fix-check-secret@127.0.0.1:2/0?token=pm-review-fix-check-secret#fragpm-review-fix-check-secret", want sanitized topology endpoint
FAIL
FAIL	polymetrics.ai/internal/runtimecheck	0.449s
FAIL
```

## Exact green outputs

```bash
gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/runtime_cli_test.go
go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	11.749s
```

```bash
go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	17.329s
```

```bash
go test ./internal/cli/...
go vet ./...
go build ./cmd/pm
```

```text
ok  	polymetrics.ai/internal/cli	195.015s
# go vet and go build emitted no output and exited 0
```

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

```text
go test ./... pass; slow packages included internal/cli (cached) and internal/connectors/certify 342.381s.
make verify pass; tail included:
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
# gofmt, go vet, and go build emitted no output and exited 0
```

```bash
./pm help runtime
./pm runtime
./pm runtime --help
./pm runtime --json
./pm runtime bogus --json
./pm runtime doctor --unknown ignored extra --root "$root" --json
```

```text
manual kind=CommandManual command=runtime bytes=600
invalid action category=usage code=usage_error stderr='error: unknown command "bogus" for "pm runtime"'
doctor kind=RuntimeDoctor statuses={'postgres': 'error', 'dragonfly': 'error', 'temporal': 'error'} stderr_bytes=0 secret_leak=0
help bytes=     470
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
Validated connector docs in docs/connectors
Wrote 11 docs pages to lib/docs.generated.ts.
# diff -ru, git diff --check, and go.mod/go.sum diff emitted no output
```

## Review-fix exact green outputs

```bash
gofmt -w internal/cli/cobra_router.go internal/cli/cobra_router_test.go internal/cli/runtime_cli_test.go internal/runtimecheck/runtimecheck.go internal/runtimecheck/runtimecheck_test.go internal/cli/docs.go
go test ./internal/cli/... -run 'Runtime|CobraRouterShell' -count=1
go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1
go test ./internal/runtimecheck/... -count=1
```

```text
ok  	polymetrics.ai/internal/cli	11.600s
ok  	polymetrics.ai/internal/cli	17.208s
ok  	polymetrics.ai/internal/runtimecheck	0.364s
```

```bash
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1
npm --prefix website run gen:docs
```

```text
ok  	polymetrics.ai/internal/cli	10.168s
Wrote 11 docs pages to lib/docs.generated.ts.
```

```bash
gofmt -w cmd internal
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
make verify
git diff --check
git diff -- go.mod go.sum
```

```text
go test -timeout 20m ./... pass; slow packages included internal/cli 222.610s, internal/connectors/certify 369.637s, internal/runtimecheck 2.323s.
make verify pass; tail included:
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
# gofmt, go vet, go build, git diff --check, and go.mod/go.sum diff emitted no output
```

```bash
TMP_DOCS=$(mktemp -d); ./pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"; diff -ru docs/cli "$TMP_DOCS/cli"; ./pm docs validate --connectors-dir docs/connectors
./pm runtime --root
./pm runtime doctor --root
./pm --json runtime --root
./pm --root "$ROOT" --json runtime doctor
./pm help runtime
./pm runtime
./pm runtime --help
```

```text
Generated docs in <tmp>/cli and connector docs in <tmp>/connectors
Validated connector docs in docs/connectors
malformed root flags exit 2 with usage category
runtime doctor endpoints sanitized; absent services reported degraded
runtime help parity ok
```
