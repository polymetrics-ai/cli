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
| 1 | RED | Add focused native ETL tree, flags, compatibility, validation, cancellation, event/output tests; run focused CLI tests | Pending |
| 2 | GREEN | Native ETL tree + typed handlers + ETL-only normalization; remove ETL wrapper/parser use | Pending |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI/app and exact-start differential | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Parity/delivery | Runtime help, temporary docs/website/generated checks, scope/dependency guards, commit/push | Pending |

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

Pending test-only checkpoint. It must fail before any production edit and will be recorded verbatim here.
