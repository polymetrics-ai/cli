# Phase 402 TDD Ledger

Issue: #402 — migrate config-shaped env reads to typed config.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-documentation` (docs caveat only if changed).

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 402 --skip-research >/tmp/gsd-plan-phase-402.prompt
scripts/gsd prompt programming-loop init --phase 402 --dry-run >/tmp/gsd-programming-loop-402.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written.
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint. |
| 1 | Red | pending | pending | Add tests proving config-file/env alias injection for named readers and exclusions. |
| 2 | Green | pending | pending | Implement minimal config injection. |
| 3 | Refactor | pending | pending | Docs caveat and cleanup while tests green. |

## Planned red tests

- `internal/runtimecheck`: `FromConfig` maps `config.Config.Runtime`; `FromEnv` honors `PM_*` alias through typed loader.
- `internal/schedule`: typed backend config selects Temporal only when explicit addr provided; crontab file injection writes temp file without raw env dependency.
- `internal/cli`: runtime doctor config-file value appears in redacted output; worker status remains byte-compatible when temporal unset but accepts explicit config file; schedule install/remove honors config-file crontab path; agent image uses config-file podman bin; RLM fake runner works from config-file key.
- `internal/worker`: typed submitter activities use injected podman image/bin; cancellation tests remain green.
- Exclusion guard: credential `--from-env` and provider API-key env paths remain raw env only.

## Exact red outputs

Pending.
