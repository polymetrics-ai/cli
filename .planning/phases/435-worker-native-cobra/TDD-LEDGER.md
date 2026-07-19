# Phase 435 TDD Ledger

Issue: #435 — nativize hidden worker namespace.
Invocation: `issue-435-pi-sol-high-20260719T064417Z`
Model/thinking profile: `Sol` / `high`
Starting HEAD: `14c02d295065c3bf33c65eaac5f8d36642798f81`

## GSD and skills

Doctor/list passed; `plan-phase 435 --skip-research` generated and is executed inline. The adapter lacks `programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity and exact start before tests or production edits | Complete |
| 1 | RED | `go test ./internal/cli -run 'TestWorker(Command|Status|Help|Invalid|Config|Context)' -count=1` | Failed as required before production edits: undefined `newWorkerCobraCommand`, `workerCommandRuntime`, and `newRootCmdWithWorkerRuntime` |
| 2 | GREEN | Native hidden worker subtree + invocation-local fake runtime + worker-only dispatcher removal | Pass: focused `0.569s`; repeated ×5 `0.738s`; focused race `1.690s` |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI, worker fake, config/parity/differential checks | Pass: full CLI `427.774s`; worker/config fakes pass; exact-start 6/6 compatible plus 2/2 intentional help changes |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` default-only | Pass: full repository CLI `435.094s`, certify `344.412s`; `make verify` exited 0 with docs/smoke/lint/connectors green |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Original delivery completed, but P2 invalidated the fake-only evidence claim |
| 6 | P2 correction plan | At exact HEAD `f692225ab53a3c0467d42c0ac3e9810107d73a82`, accept `/tmp/pm-397-review-435.log`: config-migration worker status used `Run` and production `temporalprobe.Probe` | In progress; verification reset pending before test edits |
| 7 | Correction RED | Add a fake status call/address assertion while retaining `Run`; it must fail because the injected seam is bypassed | Pending |
| 8 | Correction GREEN | Route the config-migration case through `runWorkerInvocation`; preserve config/address/status assertions | Pending |
| 9 | Correction gates | Focused/repeated/race worker/config tests, network-dial audit, conditional full CLI, diff/gofmt/vet | Pending |

## RED contract

- `worker` and `status`/`serve` are native Cobra commands, while `worker` remains hidden from root discovery and absent from legacy wrappers.
- Worker has no local pflags. Current global root/json/plain/no-input/progress forms retain invocation precedence and placement compatibility.
- Bare worker plus direct long/short, `pm help worker`, positional `worker help`, and trailing action help render contextual text/JSON help and exit 0 without invoking status or serve.
- `status` and `serve` preserve trailing operand tolerance; invalid first actions remain usage errors. Unknown/malformed flags, operands, and literal `--` cannot cause later-action discovery.
- Injected fakes prove status probe and serve routing, explicit config selection, activity configuration, ready ordering, context propagation, and cancellation without Temporal, Podman, network, database, or runtime services.
- Help, invalid actions/globals/config, and missing explicit Temporal address invoke neither status nor serve.
- Text/JSON status and serve outputs, one-envelope errors, exit categories, stdout/stderr, primary-env/legacy-env/file precedence, and unrelated-config nondisclosure are pinned.
- The worker remains a typed RLM Temporal workflow host; no generic runner/action/viewer is introduced.

## RED evidence

The complete test-only contract preceded production edits. Focused CLI compilation failed on the intentionally absent native worker constructor and invocation-local runtime seam:

```text
internal/cli/worker_native_cobra_test.go:22:9: undefined: newWorkerCobraCommand
internal/cli/worker_native_cobra_test.go:311:39: undefined: workerCommandRuntime
internal/cli/worker_native_cobra_test.go:312:9: undefined: workerCommandRuntime
internal/cli/worker_native_cobra_test.go:348:9: undefined: newRootCmdWithWorkerRuntime
internal/cli/worker_native_cobra_test.go:372:9: undefined: newRootCmdWithWorkerRuntime
FAIL\tpolymetrics.ai/internal/cli [build failed]
```

The tests cover the hidden native tree, no local flag surface, status/serve fakes, bare/text/JSON/positional/trailing help, literals, malformed and legal unknown flags, strict action discovery, globals, config precedence and nondisclosure, context cancellation, output contracts, and the typed non-generic worker boundary. No production code or service was invoked.

## GREEN / refactor evidence

Native Cobra now owns hidden `worker`, `status`, `serve`, and positional help. Worker has no local pflags; worker-only normalization preserves strict first-action ownership, ignored selected-action operands, literals, malformed/unknown tokens, and makes direct/trailing help side-effect free. An invocation-local runtime carries probe and serve functions; no mutable global seam remains. Typed handlers preserve explicit Temporal config, Podman activity settings, readiness, status/serve text and JSON, task queue, error mapping, and cancellation.

Focused tests passed in `0.569s`; repeated ×5 in `0.738s`; focused race in `1.690s`; worker package tests/race in `0.614s`/`1.580s`; config tests in `0.649s`. Router/golden/docs focus passed in `6.115s`; full CLI passed in `427.774s`. Exact-start binary differential matched 6/6 unchanged status/serve-error/invalid/operand cases and confirmed 2/2 intentional contextual-help changes. Runtime help/topic/bare/direct routes match, worker remains hidden from root help, generated `docs/cli/worker.md` is present, and website docs generation has no tracked delta.

Correction: the preceding no-dial statement was false at the original verified head. `TestWorkerStatusUsesExplicitConfigFileTemporalAddr` reached production `temporalprobe.Probe` through `Run` and attempted a loopback Temporal dial. P2 is accepted; final evidence will distinguish the corrected fake-only worker/config tests from intentional Temporal dial unit tests outside that focused set.

## Final verification evidence

- `gofmt -w cmd internal`, `go vet ./...`, and `go build ./cmd/pm` passed.
- `go test -timeout 20m ./...` passed; CLI `435.094s`, certify `344.412s`, all packages green.
- Default-only `make verify` exited 0: tidy diff clean, vet/full tests/build/docs validation passed, established temporary-root smoke passed, lint reported 0 issues, and connector validation checked 547 bundles with 0 findings.
- Runtime help topic/bare/direct and status/no-config checks passed; hidden worker did not appear in root help.
- Generated CLI manual parity passes; website docs generation produced no tracked delta.
- Scope/dependency guards pass: no go.mod/go.sum, connector definition, unrelated namespace, or broad website delta; worker dispatcher and worker `parseFlags` call sites are absent; dynamic connector parsing remains.

No optional integration mode was enabled. `make verify` ran only its established dependency-free temporary-root smoke and retained reverse ETL plan → preview → approval → execute order. This historical full-gate result does not satisfy the pending P2 correction verification.
