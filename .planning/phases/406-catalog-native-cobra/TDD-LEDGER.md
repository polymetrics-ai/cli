# Phase 406 TDD Ledger

Issue: #406 — nativize catalog namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`.

Rule anchors:

- `golang-how-to`: CLI task routes to `golang-spf13-cobra` + `golang-cli`; tests route to `golang-testing`; args/I/O route to `golang-security` + `golang-safety`.
- `golang-cli`: exit codes, stdout/stderr discipline, CLI unit testing, common mistakes around direct stdout and noisy usage.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior not implementation-only.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context where propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags reference StringArray vs StringSlice and NoOptDefVal use.
- `golang-security`: trust-boundary questions #1-#3, no secrets, command args untrusted.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 406 --skip-research >/tmp/gsd-plan-phase-406.prompt
scripts/gsd prompt programming-loop init --phase 406 --dry-run >/tmp/gsd-programming-loop-406.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-406.prompt`.
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/cli/ -run 'Catalog|CobraRouterShell' -count=1` | Pending | Add focused catalog-native tests first. |
| 2 | Green | `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1` | Pending | Native catalog parser should pass with golden diff empty. |
| 3 | Refactor | `gofmt -w cmd internal` + focused gates | Pending | Keep scope minimal. |
| 4 | Gate | Full issue verification | Pending | Record exact outputs before handoff. |

## Planned red tests

- `TestCatalogCommandIsNativeCobraSubtree`: current wrapper should fail because `catalog.DisableFlagParsing` is true, no `refresh`/`show` subcommands exist, and no native `--connection` flag metadata exists.
- `TestCatalogInvalidActionIsUsageBeforeConnectionValidation`: current legacy handler should fail because `pm catalog bogus --json` reports missing `--connection` as runtime error instead of usage exit 2.

## Exact red outputs

Pending; fill immediately after red tests are run.

## Exact green outputs

Pending.
