# Phase 434 TDD Ledger

Issue: #434 — nativize RLM namespace.
Invocation: `issue-434-pi-sol-high-20260719T053630Z`
Model/thinking profile: `Sol` / `high`
Starting HEAD: `2ac457a163cbd7bc9a3708da88b03d375ec5e952`

## GSD and skills

Doctor/list passed; `plan-phase 434 --skip-research` generated and is executed inline. The adapter lacks `programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

## RED / GREEN / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity and exact start before tests or production edits | Complete |
| 1 | RED | Focused native RLM tree/routing/parser/output tests | Pending |
| 2 | GREEN | Native RLM subtree + typed handler + injected analyzer factory + RLM-only normalization; remove legacy wrapper/parser | Pending |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI, RLM, worker fake, parity and differential checks | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Pending |

## RED contract

- Native `rlm` owns current `run` and hidden positional help; no RLM legacy wrapper remains.
- Native definitions cover `spec`, `in`, `out`, `mode`, `dry-run`, and `request`, preserving repeated last-value, bare `true`, assigned/space values, unknown tolerance, ignored operands, global placement, and exact old parser edge behavior.
- Bare namespace plus `pm help rlm`, `rlm --help`, `rlm -h`, `rlm help`, and JSON manual routes preserve the canonical RLM manual and exit 0.
- Trailing help, literal `--`, short flags, malformed/legal unknown flags, invalid actions, and action/operand discovery preserve legacy outcomes.
- Injected fake analyzers prove deterministic/fixture/model/agent routing, context propagation, close behavior, request/spec/warehouse mapping, output selection, and no external model/runtime call.
- Missing/invalid flags, missing/malformed spec, analyzer/factory/close failures, text/JSON output, stdout/stderr, one-envelope errors, request non-leakage, and no generic runner/viewer are covered.
- Agent execution remains optional and dependency-free tests use only injected fakes or the existing hermetic fake runner.

## Exact RED evidence

Pending test-only edit and focused failure before production changes.
