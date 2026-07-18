# Phase 424 TDD Ledger

Issue: #424 — nativize runtime namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-code-style`.

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

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1` | Fail | `runtime` remains a legacy wrapper and native `doctor` subcommand is missing. |
| 2 | Green | `gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/runtime_cli_test.go`; `go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1` | Pass | Native runtime parser green; bare help, JSON manual, invalid action category, unknown flags, extra args, config endpoints, and redaction preserved. |
| 3 | Refactor | `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`; `go test ./internal/cli/...`; `go vet ./...`; `go build ./cmd/pm` | Pass | Focused/golden, full internal CLI package, vet, and build green. |

## Planned red tests

- `TestRuntimeCommandIsNativeCobraSubtree`: current wrapper should fail because `runtime` remains legacy; expected native `doctor` subcommand, unknown-flag whitelist, and no-file completion seam are missing.
- `TestRuntimeDoctorNativeCobraPreservesLegacySemantics`: behavior cases cover doctor JSON, unknown flag tolerance, extra args tolerance, late global `--json`, late global `--root`, config-file endpoints, and no raw secret leakage.
- `TestRuntimeBareHelpAndInvalidActionSemantics`: bare namespace help must exit 0, `--help` must render canonical docs, JSON manual must emit `CommandManual`, and invalid action must remain usage exit 2.

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
