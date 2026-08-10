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

- `.githooks/post-commit` — POSIX `sh` hook only.
- `scripts/tests/post-commit-autopush.sh` — local-Git executable acceptance
  harness only.
- `docs/GUIDE.md` — opt-in and opt-out documentation beside the existing hook
  instructions.
- `.planning/phases/cli-autopush-commit-hook-r1/**` — delivery evidence.

## Explicit exclusions

- Do not change `core.hooksPath`, Git config, CI, server-side hooks, or any
  installer in the checked-out repository.
- Do not force-push, use a `+` refspec, add dependencies, make network requests,
  or change Go/CLI/connector behavior.
- Do not replace the existing tracked pre-commit hook.

## Design

1. Exit successfully when `PM_NO_AUTOPUSH` is set, when HEAD is detached, or when
   the branch is `main`/a locally known remote default branch. The detached child
   resolves the remote's live symbolic `HEAD` and refuses an unknown or matching
   default before it can push.
2. `prepare-commit-msg` resolves each Git operation marker with `git rev-parse
   --git-path` before Git clears it, records the current parent in worktree-local
   Git state, and clears stale records on ordinary commits. `post-commit` consumes
   a matching record before also checking active `rebase-merge`, `rebase-apply`,
   `MERGE_HEAD`, `CHERRY_PICK_HEAD`, `REVERT_HEAD`, and `BISECT_LOG` paths.
3. Select the branch remote, then `remote.pushDefault`, then an existing `origin`.
   Resolve Git's common directory, then resolve rate-limit state below
   `pm-autopush/` with `GIT_DIR=<common-dir> git rev-parse --git-path`. An atomic
   sibling lock protects the timestamp decision and replacement across linked
   worktrees; a branch timestamp younger than 600 seconds (or from the future)
   skips the attempt.
4. Persist the attempt timestamp before scheduling the child. Launch `git push --
   <remote> refs/heads/<branch>:refs/heads/<branch>` with `nohup`, standard input,
   output, and error detached from the commit terminal. The refspec has no `+` and
   no force option. A child failure or default-resolution refusal adds one short
   line to the per-branch local log; the parent hook exits zero in every path.
5. The harness uses temporary bare remotes and hook wrappers to prove stale remote
   default refusal, concurrent linked-worktree leasing, real manual merge/cherry-
   pick/revert completions, a real two-commit rebase, expiry behavior, commit
   non-blocking behavior, and a genuine non-fast-forward rejection without force.

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
