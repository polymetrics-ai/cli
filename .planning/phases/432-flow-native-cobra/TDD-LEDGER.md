# Phase 432 TDD Ledger

Issue: #432 — nativize flow namespace.
Invocation: `issue-432-pi-openai-codex-gpt-5.6-sol-high-20260719T034344Z`
Model/thinking: `openai-codex/gpt-5.6-sol` / `high`
Starting HEAD: `ec12c1729e0aaf233a853eff5c6291885f910b15`

## GSD and skills

Doctor/list passed; `plan-phase 432 --skip-research` generated and is executed inline. The adapter lacks `programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

## RED / GREEN / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity and exact start before tests or production edits | Complete |
| 1 | RED | Add focused native flow tree, parser/output, cancellation/events/telemetry/checkpoint/ledger tests; run focused command and capture exact failure | Pending |
| 2 | GREEN | Native flow subtree + typed handlers + flow-only normalization/state; remove legacy wrapper/parser | Pending |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI and flow package gates, parity/differential checks | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pending |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Pending |

## RED contract

- Native `flow` owns `plan`, `preview`, `run`, `status`, `list`, and hidden positional `help`; no flow legacy wrapper remains.
- Native definitions cover all current local flags (`file`, `force`, `flows-dir`) with exact repeated/bare/assigned and ignored-unknown behavior where applicable, while global root/json/plain/no-input/progress placement and assigned booleans remain unchanged.
- Bare namespace and `pm help flow`, `flow --help`, `flow -h`, `flow help`, and JSON manual routes preserve the canonical flow manual and exit 0.
- Trailing help, literal `--`, short flags, malformed unknown flags, and legal unknown flags preserve legacy outcomes rather than becoming accidental Cobra controls.
- Invalid actions remain usage exit 2. Leading unknown/help-like/literal tokens cannot discover or execute a later valid action.
- Named run and status preserve first positional ownership. Plan/preview/run preserve file path semantics, relative RLM spec resolution, named-flow directory resolution, and all existing flow directory defaults.
- Text and JSON plan/preview/run/list/status output remains deterministic and terminal-safe; JSON invocations retain one stdout envelope and diagnostics/events remain on stderr.
- Usage errors remain exit 2; malformed manifests and invalid global/progress inputs retain validation/runtime taxonomy established by current behavior.
- Cancellation reaches the active flow adapter, stops later steps, releases the lease, emits sanitized failed lifecycle evidence, records failed ledger/telemetry evidence without sensitive values, and does not checkpoint the cancelled step.
- Successful and resumed runs retain deterministic events, telemetry spans/metrics, checkpoint skip/force behavior, and ledger running/success ordering.
- Tests use temporary manifests/project roots and in-memory fakes only; no action step, external write, credential, service, or connector request is executed.

## Evidence

RED, GREEN, refactor, parity, and final gate evidence will be appended as commands execute. No production edit is allowed before the focused RED is captured.
