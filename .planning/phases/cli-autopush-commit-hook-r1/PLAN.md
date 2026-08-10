# PLAN — Time-boxed rebase-safe post-commit push

## GSD setup and fallback

- `scripts/gsd doctor` passed in this worktree.
- Command sources were resolved for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`; `go run
  ./cmd/agentcontractgen check` passed.
- This task is not a numbered roadmap phase, the Pi runtime is unavailable here,
  and the task requires one autonomous worker. The documented manual inline GSD
  fallback is used. This plan, the TDD ledger, verification checklist, and final
  review record are the lifecycle evidence.
- Required skill routing was reviewed. `no-mistakes` was loaded for the later
  validation/PR gate. No Go, `pm` command surface, connector, website, or runtime
  behavior is changed, so Go/CLI/website task-specific skills are not applicable.

## Goal

Let an operator who explicitly enables the tracked hooks preserve committed
feature-branch work remotely within ten minutes, without pushing rebase replay
commits, default/detached heads, or any non-fast-forward update.

## Allowed implementation scope

- `.githooks/post-commit`, `.githooks/prepare-commit-msg`, and
  `.githooks/pm-autopush-operation-state` — POSIX `sh` hook support only.
- `scripts/tests/post-commit-autopush.sh` — local-Git executable acceptance
  harness only.
- `docs/GUIDE.md` — opt-in and opt-out documentation beside the existing hook
  instructions.
- `.planning/phases/cli-autopush-commit-hook-r1/**` — delivery evidence.

## Explicit exclusions

- Do not change `core.hooksPath`, Git config, CI, server-side hooks, or any
  installer in the checked-out repository.
- Do not force-push, use a `+` refspec, add dependencies, make synchronous
  network requests from the commit path, or change Go/CLI/connector behavior.
- Do not replace the existing tracked pre-commit hook.

## Design

1. Exit successfully when `PM_NO_AUTOPUSH` is set, when HEAD is detached, or when
   the branch is `main`/a locally known remote default branch. The detached child
   resolves every effective push URL's live symbolic `HEAD` and refuses an unknown
   or matching default before it can push.
2. A shared sourced helper resolves every Git operation path with `git rev-parse
   --git-path`, including sequencer, squash, and merge message state.
   `prepare-commit-msg` records a commit-scoped marker before Git clears that state
   and clears stale records on ordinary commits. `post-commit` consumes that marker
   before using the same helper in both parent and detached child paths.
3. Select `branch.<name>.pushRemote`, then `remote.pushDefault`, then
   `branch.<name>.remote`, falling back to an existing `origin` only when none
   is configured. Resolve Git's common directory, then resolve the local log below
   `pm-autopush/` with `GIT_DIR=<common-dir> git rev-parse --git-path`. The
   timestamp itself is a blob behind a shared `refs/pm-autopush/<branch>` ref;
   `git update-ref` compare-and-swap atomically records an eligible attempt across
   linked worktrees, while a younger or future timestamp skips the attempt.
4. Persist the attempt timestamp before scheduling the child with the current HEAD
   OID. The child re-resolves Git operation state, then only sends that OID after
   confirming the branch still names it, using `git push -- <remote>
   <scheduled-oid>:refs/heads/<branch>` with `nohup`, standard input, output, and
   error detached from the commit terminal. The refspec has no `+` and no force
   option. A child failure, operation, branch-change, or default-resolution refusal
   adds one short line to the per-branch local log; the parent hook exits zero in
   every path.
5. The harness uses temporary bare remotes and hook wrappers to prove stale and
   pushurl-specific default refusal, branch-specific/global/tracking push-target
   precedence, concurrent linked-worktree compare-and-swap from an expired rate
   timestamp, delayed children during and after a manual merge, clean and amended
   squash, cherry-pick, and revert completions, a real two-commit rebase, expiry behavior,
   commit non-blocking behavior, and a genuine non-fast-forward rejection without
   force.

## TDD task sequence

1. **Red:** add the POSIX harness and run it while the hook is absent; retain the
   failing output.
2. **Green:** implement the hook with all refusal, state, non-force, and detached
   push rules; rerun the harness.
3. **Regression:** run `shellcheck -s sh`, execute the harness, and inspect the
   resulting diff to prove no global hook configuration or CI path changed.
4. **Documentation:** describe the deliberately manual `core.hooksPath` opt-in
   and the environment opt-out without obscuring the existing pre-commit cost.
5. **Verification/review:** complete the local checks, manual GSD verification,
   and a focused security/safety review before committing.
