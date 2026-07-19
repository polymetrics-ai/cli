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
| 2 | GREEN | Native hidden worker subtree + invocation-local fake runtime + worker-only dispatcher removal | Pending |
| 3 | Refactor | Focused/repeated/race/router/golden/full CLI, worker fake, config/parity/differential checks | Pending |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` default-only | Pending |
| 5 | Delivery | Finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Pending |

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

Pending.
