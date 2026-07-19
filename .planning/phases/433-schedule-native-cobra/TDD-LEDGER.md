# Phase 433 TDD Ledger

Issue: #433 — nativize schedule namespace.
Invocation: `issue-433-pi-sol-high-20260719T044819Z`
Model/thinking profile: `Sol` / `high`
Starting HEAD: `ab1c79eede67fa87e1c6b808d6ddba0b27fcf00d`

## GSD and skills

Doctor/list passed; `plan-phase 433 --skip-research` generated and is executed inline. The adapter lacks `programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-documentation`, `golang-spf13-cobra`.

## RED / GREEN / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity and exact start before tests or production edits | Complete |
| 1 | RED | `go test ./internal/cli -run 'TestSchedule(Command|Create|Install|Help|Backend)' -count=1` | Failed as required before production edits: undefined `newScheduleCobraCommand`, `scheduleCommandRuntime`, and `newRootCmdWithScheduleRuntime` |
| 2 | GREEN | Native schedule subtree + typed handlers + injected runtime seam + schedule-only normalization; remove legacy wrapper/parser | Pass: focused all-schedule CLI `0.595s`; schedule package `0.598s`; repeated ×5 `0.655s`; focused race CLI `55.681s`; router/golden/schedule `6.728s`; exact-start differential 104/104 |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI and schedule package gates, parity/differential checks | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Pending |

## RED contract

- Native `schedule` owns current create/list/install/remove and hidden positional help; no schedule legacy wrapper remains.
- Native definitions cover all current local flags (`name`, `cron`, `flow`, `crontab`) with exact repeated/bare/assigned and ignored-unknown behavior where applicable, while global root/json/plain/no-input/progress placement and assigned booleans remain unchanged.
- Bare namespace and `pm help schedule`, `schedule --help`, `schedule -h`, `schedule help`, and JSON manual routes preserve the canonical schedule manual and exit 0.
- Trailing help, literal `--`, short flags, malformed unknown flags, and legal unknown flags preserve legacy outcomes rather than becoming accidental Cobra controls.
- Invalid actions remain exit-2 usage errors. `uninstall`, `run`, and `history` remain invalid under the current contract and cannot reach an install/remove backend. Leading unknown/help-like/literal tokens cannot discover or execute a later valid action.
- Install/remove preserve first positional ownership, bare/assigned/repeated `--crontab`, schedule-not-found behavior, root in rendered payload, context propagation, default backend selection inputs, non-crontab crontab-cleanup fallback, and best-effort backend removal before manifest deletion.
- Create preserves last-value flag selection, cron and manifest-name validation, conflict validation, persisted manifest timestamps, exact text/JSON shapes, and no backend calls.
- List preserves ignored tails, empty JSON slice, lexicographic manifest order, text/JSON determinism, and no backend calls.
- Usage errors remain exit 2; invalid cron/name/not-found remain validation exit 3; install failures retain internal/runtime legacy classification and wrapping.
- Tests use temporary roots, redirected temporary crontab files, fixed clocks, executable stubs, and fake backends only; no real crontab/launchd/systemd/Temporal command or scheduler service executes.

## Exact RED evidence

Captured after the complete test-only edit and before any production edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/schedule_native_cobra_test.go:23:9: undefined: newScheduleCobraCommand
internal/cli/schedule_native_cobra_test.go:441:13: undefined: scheduleCommandRuntime
internal/cli/schedule_native_cobra_test.go:450:9: undefined: newRootCmdWithScheduleRuntime
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The intentionally missing constructors/runtime prove the native tree and injected scheduler seam do not yet exist. The committed tests cover current create/list/install/remove/help, every current local flag and first operand, invalid `uninstall`/`run`/`history`, bare/text/JSON/positional/trailing help, literal/malformed unknown/action discovery/global booleans, fixed-time deterministic output, cron/name/not-found classification, context/root/backend behavior, and effect-free fake install/remove cleanup. Existing schedule removal tests were tightened to force the redirected temporary crontab backend so no platform scheduler can execute.

## Focused GREEN

`newScheduleCobraCommand` now owns create/list/install/remove/help with typed string arrays, a native boolean `--crontab`, unknown tolerance, and no-file completion seams. Typed handlers retain current cron/name/conflict validation, manifest format/timestamps, root-bearing scheduler payload, typed config selection, context propagation, install wrapping, best-effort removal plus crontab fallback, and text/JSON output. An invocation-local runtime seam injects fixed clocks, executable paths, selectors, and fake backends in tests; production defaults still use the existing backend selector and implementations.

Schedule-only first-operand capture and normalization preserve the legacy parser's space/assigned/repeated flag behavior, arbitrary false `--crontab` values, help/literal/malformed unknown tails, strict first operand, and invalid action ownership. Schedule is absent from `cobraLegacyCommands`; `runSchedule` and schedule `parseFlags` call sites are removed; dynamic connector `parseFlags` remains.

All schedule CLI tests passed in `0.595s`; schedule package tests passed in `0.598s`; repeated schedule CLI ×5 passed in `0.655s`; focused race passed for CLI in `55.681s`; router/golden/schedule focus passed in `6.728s`. A 104-case exact-start differential across create/list/install/remove and 26 action tails matched exit/stdout/stderr exactly after timestamp normalization. No real scheduler command, service, credential, or dependency was used.

## Final refactor and verification evidence

Pending.
