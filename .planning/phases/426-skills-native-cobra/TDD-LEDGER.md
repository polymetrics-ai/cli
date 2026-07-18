# Phase 426 TDD Ledger

Issue: #426 — nativize skills namespace.
Invocation session: `issue-426-pi-openai-codex-gpt-5.6-sol-high-20260718T104457Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `54bfcbab5a997c81676b286fe78de00a109f5fba`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 426 --skip-research` generated the official prompt. `scripts/gsd prompt programming-loop init --phase 426 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | Focused native skills/router tests | Pending |
| 2 | GREEN | Native namespace/action/typed flags and legacy parser removal | Pending |
| 3 | Refactor | Focused router/golden/full CLI tests | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Parity/safety | built binary, docs/website/generated/golden/dependency/scope checks | Pending |

## Planned RED coverage

- `skills` is native and absent from legacy wrappers; `generate` is a native action with a native `stringArray` `--dir`, `NoOptDefVal=true`, unknown-flag whitelist, and no-file completion fallback.
- `generate` preserves `--dir value`, `--dir=value`, repeated-last-wins, bare `--dir=true`, comma-preserving values, unknown flags, ignored extra positionals, plain output, JSON `SkillGeneration`, and missing/empty-dir validation.
- Bare/text/JSON/flag/short/positional help all resolve to the canonical skills manual.
- Invalid action remains usage exit 2 and does not render help; malformed global boolean remains validation exit 3.
- Global/config forms cover `--root value`, `--root=value`, late placement, `--json`, `--json=true`, and `--json=false` overriding config/env JSON.
- Existing generation contract writes only expected metadata skill files under test temp directories and contains no secret values.

Exact RED and GREEN outputs will be appended after execution; no production file will be edited before RED is captured.
