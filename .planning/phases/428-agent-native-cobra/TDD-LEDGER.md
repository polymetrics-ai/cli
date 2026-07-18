# Phase 428 TDD Ledger

Issue: #428 — nativize agent namespace.
Invocation session: `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `235233f7cfde4a24612be6b0f95fb37a412d388a`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 428 --skip-research` generated 10668 bytes and is being executed inline. `scripts/gsd prompt programming-loop init --phase 428 --dry-run` failed because the command is absent, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-documentation`, `golang-spf13-cobra`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | Focused agent/router tests after test-only edits | Pending |
| 2 | GREEN | Native tree, typed request, agent-only compatibility, injected image runtime, validation | Pending |
| 3 | Refactor | Focused router/golden/full CLI and legacy differential | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Parity/safety | built binary help/plan only; docs/website/generated/runtime dependency-free/scope guards | Pending |

## Planned RED coverage

- `agent` is native and absent from legacy wrappers; native `plan`, `image`, `image build|pull|ensure`, and hidden positional `help` nodes exist.
- `plan --request` is `stringArray`, bare value is `true`, unknown flags remain tolerated, no-file completion seam exists, repeated values remain last-wins.
- Plan spaced/assigned/bare/repeated/unknown/extra-positional forms preserve deterministic exact text and JSON output.
- Bare/text/JSON/long/short/positional help routes preserve the canonical manual; invalid agent/image actions remain usage errors.
- Global root/json/plain/no-input assigned booleans and configured JSON override behavior remain stable.
- Every image action runs only against an injected fake: build command/context/Containerfile path, pull, ensure-present, ensure-pull, exact text/JSON kinds/status, error propagation, and deterministic invocation ordering.
- Request/control, project/build path, Podman binary, and image reference validation reject unsafe input before external execution.
- Legacy action-tail `--help`/`-h` and continuation after literal `--` remain compatible without Phase 19 help churn.

## Exact RED

Pending test-only checkpoint. No production file will be edited before this section records the focused failure.
