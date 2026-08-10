# PLAN — unpushed-work safety net r1

## GSD lifecycle and execution mode

- `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check`: passed before planning.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: resolved
  through the installed adapter.
- `scripts/gsd prompt discuss-phase cli-unpushed-work-safety-net-r1 --auto` and
  `scripts/gsd prompt plan-phase cli-unpushed-work-safety-net-r1 --tdd --skip-research` were
  generated and executed inline. The domain is bounded by the preserved failed-hook report, so
  skipping separate research is intentional.
- `scripts/gsd prompt execute-phase cli-unpushed-work-safety-net-r1 --no-cross-ai` was generated
  and executed inline after the red harness was recorded. This is the adapter-supported local
  execution flag; no worker role was spawned.
- Inline/manual execution is the compatible fallback because this task's canonical single-worker
  rules prohibit role spawning. It does not weaken the required
  `discuss-phase → plan-phase --tdd → execute-phase → verify-work → code-review` evidence.
- `programming-loop` is not invoked: the installed adapter does not provide it and repository
  rules explicitly prohibit fabricating it.

## Required skills

- `no-mistakes` — final branch validation/PR workflow only; do not start or disturb the shared
  daemon until firstmate asks for that gate.
- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and
  `gsd-code-review` — required lifecycle artifacts and review execution.
- No Go/CLI/connector/website skill applies: this is a Python standard-library developer safety
  utility, not a `pm` command or product/website surface. CLI help/manual/website parity is N/A.

## Slice 1 — red acceptance harness

1. **RED:** add an executable Python integration test that creates real bare Git remotes, clones,
   linked worktrees, commits, a real multi-commit conflicted rebase, divergent remote history, and
   a live `fcntl` lock. It must fail because the observer script does not exist.
2. **GREEN:** implement the smallest standard-library observer that can be opt-in enabled, scans
   real worktrees, reports all refusal states, and pushes only an explicit feature-branch commit
   SHA with no force mode.
3. **REFACTOR:** make event names and persistent state deterministic enough for the test and for
   human `status` output; add no third-party package.

## Slice 2 — safety, recovery, and operator path

1. **RED:** expand the harness to assert real recovery after dirty work is committed and after a
   genuine multi-commit rebase completes; assert a divergent remote remains unchanged.
2. **GREEN:** add effective push-target/default discovery, current-operation/ref rechecks,
   rate-floor state written before a push, kernel advisory locking, durable errors, and a launchd
   plist generator with explicit enable/disable commands.
3. **REFACTOR:** document the exact opt-in, status, and opt-out commands; state why the existing
   pre-commit hook remains untouched.

## Safety contracts

- No `--force`, `+<refspec>`, ref rewrite, `reset`, or destructive cleanup path is permitted.
- Never push a branch whose default status cannot be confirmed from its effective push destination.
- Never push default, detached, dirty, active-operation, rate-limited, unconfigured, unknown, or
  diverged worktrees. Each case emits an explicit event.
- Use a kernel-released lock rather than a filesystem lease. Lock contention exits visibly and a
  process crash releases the lock; state read/write failure fails closed and visibly.
- The final recheck pins `refs/heads/<branch>` to a SHA before the non-force push, avoiding a
  mutable `HEAD` refspec. A concurrent rewrite can only cause a safe rejection or a push of the
  previously settled SHA, never a remote rewrite.

## Verification plan

- `python3 scripts/tests/unpushed-work-safety-net_test.py` — all refusal and recovery paths using
  real Git operations, including multi-commit rebase and no-rewrite divergence proof.
- `python3 scripts/unpushed-work-safety-net.py --help` and generated launchd plist inspection.
- `python3 -m py_compile scripts/unpushed-work-safety-net.py
  scripts/tests/unpushed-work-safety-net_test.py`.
- `make unpushed-work-safety-net-check`, `git diff --check`, and relevant independent Make gates.
- `scripts/gsd prompt verify-work cli-unpushed-work-safety-net-r1 --auto` and
  `scripts/gsd prompt code-review cli-unpushed-work-safety-net-r1` after implementation;
  record their inline/manual execution in `VERIFICATION.md` and `SUMMARY.md`.

## Commit checkpoints

1. Planning evidence before production edits.
2. Red harness checkpoint if useful for review; it will remain local until firstmate starts the
   no-mistakes delivery gate.
3. Green implementation, docs, and verification checkpoint.
4. Review-fix checkpoint only if the required review discovers an in-scope issue.
