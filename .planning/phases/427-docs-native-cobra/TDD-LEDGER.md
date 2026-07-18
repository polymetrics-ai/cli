# Phase 427 TDD Ledger

Issue: #427 — nativize docs namespace.
Invocation session: `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `ab847da28ebf78e5732ac1edcde8e39f92dc2656`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 427 --skip-research` generated the official prompt and it was executed inline. `scripts/gsd prompt programming-loop init --phase 427 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | Focused docs/router tests after test-only edits | Pending |
| 2 | GREEN | Native namespace/actions/typed flags and docs-only legacy parser removal | Pending |
| 3 | Refactor | Focused docs/router/golden + full CLI tests | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Parity/safety | built binary, docs/website/generated/golden/dependency/scope checks | Pending |

## Planned RED coverage

- `docs` is native and absent from legacy wrappers; `generate` and `validate` are native actions with declared `stringArray` output-directory flags, `NoOptDefVal=true`, unknown-flag whitelists, and no-file completion fallback.
- `generate` preserves `--dir` and `--connectors-dir` spaced/assigned/repeated-last-wins/bare/comma forms, exact output text, exact CLI-manual bytes, and expected connector artifacts.
- `validate` preserves `--connectors-dir`, fallback `--dir`, default semantics where safely applicable, unknown flags, extra positionals, and exact output text.
- Bare/text/JSON/flag/short/positional help all resolve to the canonical docs manual.
- Invalid action remains usage exit 2; missing/empty generate dir and malformed assigned globals retain existing categories; errors are not masked by help.
- Global/config forms cover `--root value`, `--root=value`, late placement, assigned `--json=true|false`, `--plain=false`, and `--no-input=true`.
- Filesystem tests write only under safe temporary roots, assert all generated relative paths remain local, compare CLI docs byte-for-byte to the canonical map/checked-in tree, and validate connector output without secrets.

## Exact RED

Pending test-only checkpoint. Production files remain untouched until this section records the failing focused command and output.

## Focused GREEN

Pending.

## Full GREEN and parity

Pending.
