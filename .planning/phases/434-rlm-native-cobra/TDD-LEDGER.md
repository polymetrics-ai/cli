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
| 1 | RED | `go test ./internal/cli -run 'TestRLM(Command|Run|Help)' -count=1` | Failed as required before production edits: undefined `newRLMCobraCommand`, `rlmCommandRuntime`, and `newRootCmdWithRLMRuntime` |
| 2 | GREEN | Native RLM subtree + typed handler + injected analyzer factory + RLM-only normalization; remove legacy wrapper/parser | Pass: focused RLM/build-agent `0.582s`; repeated ×5 `8.317s`; focused CLI race `1.681s`; RLM packages `0.750s`/router `0.389s`; worker fakes `0.572s`; router/golden focus `7.918s`; RLM/worker fake CLI `1.891s`; RLM+worker race `1.718s`; exact-start differential 24/24 |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI, RLM, worker fake, parity and differential checks | Pass: full CLI `424.801s`; final RLM/worker `0.242s`/`0.358s`/`0.645s`; final golden/router/help `7.235s`; parity unchanged |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass: gofmt/vet/build; full repository CLI `431.499s`, certify `339.553s`; `make verify` `24.201s` cached full gate |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Complete after terminal artifact commit/push |

## RED contract

- Native `rlm` owns current `run` and hidden positional help; no RLM legacy wrapper remains.
- Native definitions cover `spec`, `in`, `out`, `mode`, `dry-run`, and `request`, preserving repeated last-value, bare `true`, assigned/space values, unknown tolerance, ignored operands, global placement, and exact old parser edge behavior.
- Bare namespace plus `pm help rlm`, `rlm --help`, `rlm -h`, `rlm help`, and JSON manual routes preserve the canonical RLM manual and exit 0.
- Trailing help, literal `--`, short flags, malformed/legal unknown flags, invalid actions, and action/operand discovery preserve legacy outcomes.
- Injected fake analyzers prove deterministic/fixture/model/agent routing, context propagation, close behavior, request/spec/warehouse mapping, output selection, and no external model/runtime call.
- Missing/invalid flags, missing/malformed spec, analyzer/factory/close failures, text/JSON output, stdout/stderr, one-envelope errors, request non-leakage, and no generic runner/viewer are covered.
- Agent execution remains optional and dependency-free tests use only injected fakes or the existing hermetic fake runner.

## Exact RED evidence

Captured after the complete test-only edit and before any production change:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/rlm_native_cobra_test.go:23:9: undefined: newRLMCobraCommand
internal/cli/rlm_native_cobra_test.go:434:38: undefined: rlmCommandRuntime
internal/cli/rlm_native_cobra_test.go:435:9: undefined: rlmCommandRuntime
internal/cli/rlm_native_cobra_test.go:457:9: undefined: newRootCmdWithRLMRuntime
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The missing constructor/runtime symbols prove the native RLM tree and injected analyzer seam do not exist. The tests cover every current local flag; all four mode routes through fakes; context/closer/spec/warehouse mapping; bare/text/JSON/positional/trailing help; literal/malformed/unknown/action/operand discovery; globals; stdout/stderr; request non-leakage; invalid generic/viewer actions; and malformed spec/error paths without any model, Temporal, Podman, or service call.

## Focused GREEN

Native Cobra now owns `rlm run` and hidden positional help. Six current flags are `StringArrayVar` definitions with `NoOptDefVal=true`; RLM-only normalization retains the old parser's repeated/space/assigned/bare, unknown, malformed, literal, help, short-token, operand, and strict first-action behavior. A typed run handler preserves validation/spec parsing/warehouse/dry-run/output/error behavior. An invocation-local analyzer factory routes deterministic, fixture, model, and agent modes; tests inject fakes for every route and verify context/closer/request handling without external calls. Only the RLM legacy wrapper, dispatcher, and `parseFlags` use were removed.

Focused RLM/build-agent tests passed in `0.582s`; repeated ×5 in `8.317s`; focused CLI race in `1.681s`; RLM packages in `0.750s` plus router `0.389s`; worker fake focus in `0.572s`; router/golden focus in `7.918s`; RLM/worker fake CLI in `1.891s`; RLM+worker package race in `1.718s`. An exact-start binary differential matched 24/24 help, required/unknown mode, model stub, deterministic/fixture dry-run, repeated/space/assigned/bare, unknown/malformed/literal/trailing-help, invalid-action, and global-flag cases after normalizing duration only.

## Final refactor and verification evidence

- Full `go test -timeout 15m ./internal/cli/... -count=1` passed in `424.801s`.
- Final `go test ./internal/rlm/... ./internal/worker/... -count=1` passed (`0.242s`, `0.358s`, `0.645s`); focused golden/router/RLM help passed in `7.235s`.
- Runtime help topic/bare/long/short/positional routes are byte-equal; JSON manual matches canonical text; invalid `viewer` remains exit-2 usage.
- Temp-generated `docs/cli/rlm.md` matches checked-in docs; website docs generation produced no tracked delta; golden fixture is unchanged.
- `gofmt -w cmd internal`, `go vet ./...`, and `go build ./cmd/pm` passed.
- `go test -timeout 20m ./...` passed (CLI `431.499s`, certify `339.553s`, all packages green).
- `make verify` passed in `24.201s` with cached full tests, docs validation, established temporary-root local smoke, lint 0 issues, and connector validation 547/0.
- Scope/dependency guards pass: no go.mod/go.sum, connector definition, docs/website/golden, generated, or unrelated namespace delta; no `runRLM` dispatcher or RLM `parseFlags` call remains; dynamic connector `parseFlags` remains.

No model, Temporal endpoint, Podman command, worker service, optional runtime service, live credential, secret/request value, new dependency, generic runner, PR, or review was used. `make verify` ran only its established temporary local sample smoke and preserved reverse ETL plan → preview → approval → execute order.
