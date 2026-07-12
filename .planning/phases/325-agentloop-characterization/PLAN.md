# Plan: Phase 0 Incident Replay and Migration Fuse

## Preconditions and workflow evidence

- Issue #325 and parent issue #323 are the task contracts; draft parent PR #324 exists.
- This worker is parent-spawned in an isolated worktree. Parent branch artifacts are read-only.
- `scripts/gsd doctor` passes with Node 24.
- `scripts/gsd prompt programming-loop init --phase 325-agentloop-characterization --dry-run`
  fails because the adapter registry omits `programming-loop`; this is a recorded manual-GSD
  fallback.
- Installed `gsd-programming-loop` helper preflight ran. Its generic UI output was removed because
  it was outside issue scope. The phase-local scaffold is retained and corrected here.
- Skills loaded before edits: `gsd-programming-loop`, `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-naming`.

## Slice boundaries

- [x] id:fixture-replay type:behavior — add thirteen strict synthetic fixtures, typed decoding,
  semantic replay rules, expectation matching, deterministic ordering, and stable validation codes.
- [x] id:structural-redaction type:behavior — reject raw prompt/command/session-path surfaces and
  secret-shaped strings through one structural scanner; prove with runtime-constructed canaries.
- [x] id:safety-policy type:behavior — add immutable closed safety status, tracked entrypoint
  inventory, and typed guard denial with no enable route.
- [x] id:loopctl-cli type:behavior — add replay and safety command wiring with tested help,
  stdout/stderr, JSON, and exit-code contracts.
- [x] id:driver-fuse type:behavior — guard both tracked drivers before persistence or launch while
  retaining help.
- [x] id:entrypoint-enumeration type:behavior — add shell tests that compare marked driver wrappers
  to the canonical inventory and prove run/resume cannot invoke harmless stubs.
- [x] id:make-gate type:behavior — add `agent-loop-test` and include it in `make verify` without
  weakening any existing target.
- [x] id:phase-memory type:docs — keep GSD ledger, verification, summary, prompts, run state, role
  traces, and worker handoff synchronized with actual evidence.

## Strict TDD checkpoints

1. Planning checkpoint: commit only issue-scoped phase artifacts.
2. Red checkpoint: add tests plus all thirteen fixtures, run targeted Go and shell commands, record
   expected failures, and run `tdd-gate.mjs`; no production code exists yet.
3. Green checkpoint: implement the smallest standard-library slice and make targeted/race/shell
   gates pass.
4. Refactor/review checkpoint: simplify only after green, run full verification, update all
   evidence, and commit any review fixes separately.

## Verification order

1. `go test ./internal/agentloop/... -count=1`
2. `go test -race ./internal/agentloop/... -count=1`
3. `go test ./cmd/loopctl/... -count=1`
4. `bash scripts/tests/auto-loop-control.sh`
5. `make agent-loop-test`
6. `make verify`
7. `git diff --check`

## Scope audit

Before every commit, compare `git diff --name-only` and untracked files to the issue write scope.
The child branch alone may be pushed. The worker may open one stacked PR against
`fix/323-auto-loop-hardening`; it may not merge or modify parent PR #324.
