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
| 1 | RED | `go test ./internal/cli ./internal/flow -run 'TestFlow(Command|Plan|Help|Exact)|TestEngineCancellationPreserves' -count=1` | Failed as required before production edits: `internal/cli/flow_native_cobra_test.go:20:9: undefined: newFlowCobraCommand`; flow cancellation/events/telemetry/checkpoint/ledger contract passed in `0.394s` |
| 2 | GREEN | Native flow subtree + typed handlers + flow-only normalization; remove legacy wrapper/parser | Pass: focused native contract and cancellation package `5.002s`; all `TestFlow*` CLI `5.742s`; flow package `0.301s` |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI and flow package gates, parity/differential checks | In progress: router/golden/flow `13.293s`; repeated ×5 `27.066s`; focused race CLI `60.885s`, flow `1.348s` |
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

## Exact RED evidence

Captured after the complete test-only edit and before any production edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/flow_native_cobra_test.go:20:9: undefined: newFlowCobraCommand
FAIL\tpolymetrics.ai/internal/cli [build failed]
ok  \tpolymetrics.ai/internal/flow\t0.394s
FAIL
```

The intentionally missing constructor proves the native tree does not yet exist. The committed tests cover all current actions/flags and first operands; bare/text/JSON/positional help; trailing help/literal/malformed unknown/action discovery/global booleans; local temp plan/preview/run/list/status; deterministic output and exact exits; and cancellation with lifecycle events, redacted telemetry, checkpoint absence, ledger ordering, lease release, and no later-step execution. No production file, external write, action step, credential, service, or dependency was used.

## Focused GREEN

`newFlowCobraCommand` now owns plan/preview/run/status/list/help with typed file/directory arrays, a native bare boolean force flag, unknown tolerance, no-file completion seams, and flow-only legacy control/malformed normalization. Typed handlers retain current manifest, DAG, relative-spec, named-run, directory, checkpoint, event, telemetry, and output behavior. Flow is absent from `cobraLegacyCommands`; `runFlow` and `parseFlowFlags` are removed; dynamic connector `parseFlags` remains.

Focused native/cancellation tests passed in `5.002s`; every `TestFlow*` CLI test passed in `5.742s`; all flow package tests passed in `0.301s`. Repeated ×5 passed in `27.066s`; focused race passed for CLI in `60.885s` and flow in `1.348s`; router/golden/flow focus passed in `13.293s`. No action step, external write, service, credential, or dependency was used.

Broader differential, parity, full CLI/repository, formatting, vet, build, and final verification remain pending.
