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
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI and flow package gates, parity/differential checks | Pass: post-correction race CLI `184.256s`, flow `1.395s`, events `1.740s`, telemetry `2.034s`; router/golden/flow ×5 `82.184s`; full CLI `430.094s`; flow/events/telemetry `0.666s`/`0.506s`/`0.418s`; parity and 200/200 differential pass |
| C1 | Correction RED | `go test ./internal/cli -run '^TestFlowLegacyParserEdgesKeepExactOutcomes$' -count=1` after the 200-case differential exposed 20 mismatches | Failed as required before correction production edits: 8/8 assertions failed across bare/assigned/flag-looking file values, short/unknown run/status operands, and bare flows-dir |
| C2 | Correction GREEN | Invocation-private legacy operand capture plus flow-only local-flag normalization | Pass: focused correction `4.347s`; all flow CLI `9.494s`; flow package `0.480s`; exact-start differential 200/200 |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass: vet `3.109s`; build `1.704s`; repository tests `7m16.265s`; `make verify` `24.598s` cached full gate |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Complete after this terminal artifact commit/push |

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

## Exact-parser correction RED

A 200-case exact-start differential (five actions × 40 tails) matched 180/200 and exposed only bounded pflag compatibility gaps. Before correction production edits, `TestFlowLegacyParserEdgesKeepExactOutcomes` failed all eight assertions: bare `--file` became `"true"`; assigned `--file=...` became active; `--file --force` lost the flag-looking value; run/status lost short and unknown-following positional operands; and bare `--flows-dir` stopped using the legacy default. The failures had exact exit/stdout/stderr evidence and no external effects.

## Exact-parser correction GREEN

Invocation-private operand capture now reproduces the old flow parser's first positional choice for run/status before pflag can consume short or unknown-following values. Flow-only normalization preserves missing bare string flags, ignores assigned string/force forms, consumes flag-looking string values, strips only legacy control tokens, and legalizes malformed unknowns. It does not expose an argv carrier or change normal flags, global parsing, action discovery, directory defaults, or other namespaces.

The focused correction passed in `4.347s`; every flow CLI test passed in `9.494s`; all flow package tests passed in `0.480s`; the exact-start differential now matches 200/200 exit/stdout/stderr cases across five actions and 40 tails.

## Final refactor and verification evidence

- Post-correction race command passed: CLI `184.256s`, flow `1.395s`, events `1.740s`, telemetry `2.034s`.
- Router/golden/flow focus repeated ×5 passed in `82.184s`; full CLI passed in `430.094s`; flow/events/telemetry packages passed in `0.666s`/`0.506s`/`0.418s`.
- Runtime help topic/bare/long/short/positional routes are byte-equal; JSON manual matches canonical text; invalid action remains exit 2 with the exact legacy message.
- Generated `docs/cli/flow.md` matches a temp generation; website docs generation produced no tracked delta; golden fixture is unchanged.
- `gofmt -w cmd internal`, `go vet ./...` (`3.109s`), `go test -timeout 20m ./...` (`7m16.265s`, CLI `431.217s`, certify `338.825s`), and `go build ./cmd/pm` (`1.704s`) pass.
- `make verify` passed in `24.598s` with cached full tests, docs validation, the established temporary-root local smoke, lint 0 issues, and connector validation 547/0.
- Scope/dependency guards pass: no `go.mod`, `go.sum`, connector definition, docs/website/golden, generated, or unrelated namespace delta; no `runFlow`/`parseFlowFlags` remains; dynamic connector `parseFlags` remains.

No action flow step, external write/service, live credential, dependency, PR, or review was used. The required `make verify` smoke used only its established temporary local sample/outbox path and no external reverse operation.
