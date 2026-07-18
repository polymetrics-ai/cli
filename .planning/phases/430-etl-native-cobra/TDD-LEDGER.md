# Phase 430 TDD Ledger

Issue: #430 — nativize ETL namespace.
Invocation session: `issue-430-pi-openai-codex-gpt-5.6-sol-high-20260718T225346Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `6c94754c58185df5aac53bd97587603c3154b1d5`

## GSD and skills

Doctor/list passed; the plan-phase prompt was generated and is executed inline. The adapter lacks `programming-loop`, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

## Planned RED / GREEN / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity and exact start before test/production edits | Complete |
| 1 | RED | `go test ./internal/cli -run 'TestETL(Command|Direct|RunStatus|BatchSize|HelpRoutes|UnknownInvalid|BareFlag|Cancellation|Progress)' -count=1` | Failed as required before production edits: `internal/cli/etl_cli_test.go:22:9: undefined: newETLCobraCommand` |
| 2 | GREEN | Native ETL tree + typed handlers + ETL-only normalization; remove ETL wrapper/parser use | Pass: focused contract `13.396s`; ETL/router focused suite `27.999s` |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI/app and exact-start differential | Pass |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass |
| 5 | Parity/delivery | Runtime help, temporary docs/website/generated checks, scope/dependency guards, commit/push | Pass; final evidence push pending |

## RED contract

- Native `etl` owns `check`, `catalog`, `read`, `run`, `status`, and hidden positional `help`; all actions use Cobra parsing and no ETL legacy wrapper remains.
- Current flags are `StringArray` (not comma-splitting), support repeated values, `NoOptDefVal=true`, assigned and spaced values, and unknown-flag tolerance.
- Repeated value semantics remain last-wins where handlers currently call `first`; repeated `--config` preserves all entries with later duplicate keys winning.
- Bare namespace and `pm help etl`, `etl --help`, `etl -h`, `etl help`, and JSON manual routes preserve the canonical ETL manual and exit 0.
- Action-tail help and literal `--` remain legacy-compatible rather than becoming accidental Cobra controls.
- Invalid actions are usage errors; leading unknown/help-like tokens cannot discover and execute a later ETL action.
- Global `--json`, `--plain`, and `--no-input` assigned booleans retain validation and placement behavior.
- Direct check/catalog/read use only built-in sample fixtures and temporary roots. Run/status use a temporary sample→warehouse connection.
- Batch size parses integers, defaults to 1000 in the app, and produces bounded flush counts for explicit small batches. Invalid integers are validation errors. Configured sync-mode cursor/primary-key requirements remain enforced.
- A canceled command context reaches ETL operations without replacement or goroutine leakage.
- Progress stays on stderr; final text or one JSON envelope stays on stdout; failures preserve the JSON error envelope plus stderr diagnostic; event and telemetry behavior remains context-driven.
- No service-backed `--runtime` execution, credentialed connector check, secret fixture, or reverse execution occurs.

## Exact RED

Captured after the complete focused test-only edit and before any production edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/etl_cli_test.go:22:9: undefined: newETLCobraCommand
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The missing native constructor is intentional. The test-only checkpoint specifies native ownership and every current flag, repeated/bare/assigned forms, direct fixture actions, configured run/status, batch bounds/default and sync validation, all manual routes, action-tail/literal compatibility, unknown/invalid/global inputs, action-discovery boundaries, cancellation, progress events, stdout/stderr separation, and one JSON envelope. No external connector, optional service, secret fixture, or reverse operation ran.

## Focused GREEN

`newETLCobraCommand` now owns check/catalog/read/run/status/help with typed `StringArray` flags, unknown tolerance, no-file completion seams, ETL-only spaced-value and legacy-tail normalization, and typed handlers. ETL is absent from legacy wrappers; `runETL`, `directConnector`, and their ETL `parseFlags` calls are removed. The focused contract passed in `13.396s`; the broader ETL/router focused suite passed in `27.999s`. No optional runtime service was contacted; repeated `--runtime=false` proves dependency-free default behavior.

## Local review correction

A local behavior review found that an invalid action followed by `--help` or `-h` was consumed by Cobra's namespace help flag and exited 0, bypassing the required invalid-action usage error. A focused test was added before correction production edits. Exact RED:

```text
--- FAIL: TestETLUnknownInvalidActionsGlobalsAndNoDiscoveryBypass
    error code=0, want 2; stdout={"kind":"CommandManual", ...}
FAIL polymetrics.ai/internal/cli
```

The correction bounds every unrecognized ETL action behind Cobra's literal separator before flag parsing. Valid namespace/action help and known-action legacy-tail handling remain unchanged. Focused correction GREEN passed in `0.558s`; the complete focused contract passed in `13.397s`; correction repeated ×5 and race passed in `1.061s`/`1.668s`; router/golden/help preservation passed in `6.749s`.

## Final GREEN / refactor evidence

- Initial focused contract `13.396s`; broader ETL/router `27.999s`; post-correction focused contract `13.397s`.
- Full focused repeated ×5 `65.438s`; full focused race `146.473s`; correction repeated/race `1.061s`/`1.668s`.
- Router/golden/generated manual gate `21.327s`; final golden/manual gate `7.125s`; no fixture or generated manual delta.
- ETL telemetry/progress contract `4.042s`; app ETL/sync-mode race `178.516s`.
- Full CLI `359.902s`; full app `29.499s`; full repository before final correction passed (CLI `360.105s`, certify `340.043s`).
- Exact start-vs-head differential: 20/20 exit/stdout/stderr matches for help, JSON manual, direct fixture actions, read, batch validation, missing run, trailing help/literal, invalid actions including trailing help, and global booleans.
- Runtime help topic/bare/long-help are byte-identical (4223 bytes; SHA-256 `23d099f62e2e5183b1dd01c0b47d75e3f3711907eca00196813a19748d112191`); JSON manual and invalid action usage pass.
- Website docs generation wrote 11 pages with no tracked delta; docs/CLI/golden generation and connector docs validation pass.
- `gofmt -w cmd internal`, `go vet ./...`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, and final `make verify` pass. Final `make verify`: CLI `356.154s`, certify `335.400s`, lint 0 issues, 547 connector definitions/0 findings.
- Scope/dependency guards pass: no `go.mod`, `go.sum`, connector definition, `docs/cli`, website, or unrelated namespace delta.

No external connector, optional runtime service, secret input, dependency, standalone reverse operation, PR, or review was used. The required final `make verify` ran its existing temporary-root local smoke, including the repository's built-in plan → preview → approval → run check.
