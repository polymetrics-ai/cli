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
| 1 | RED | Focused native worker tree/help/action/runtime/config/cancellation/nondisclosure tests | Pending |
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

Pending test-only checkpoint. Production code must not change until the focused test command fails for the intentionally absent native worker constructor/runtime seam.

## GREEN / refactor evidence

Pending.
