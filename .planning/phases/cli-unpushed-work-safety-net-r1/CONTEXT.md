# Context — unpushed-work safety net r1

## Locked decisions

- Replace the abandoned `post-commit` auto-push idea with an opt-in, out-of-band observer.
  It runs independently of Git hooks and scans settled worktrees periodically; it must not set
  `core.hooksPath` or change the existing pre-commit behaviour.
- Implement the observer as a dependency-free Python 3 script using only the standard library.
  Python's kernel-released advisory lock avoids a durable lease and is available on the supported
  macOS/Linux developer hosts.
- The observer is one foreground, one-shot process suitable for an explicit 600-second launchd
  schedule. Its shared state lives under the repository's Git common directory, so linked
  worktrees share one rate limit and status record.
- Auto-push only a clean, attached, non-default feature branch with a known effective push remote,
  exactly one push URL, a live-discovered remote default branch, and a fast-forward relationship
  from that push destination. The source ref is pinned to the observed commit SHA; the push uses
  an explicit non-force refspec.
- A currently active rebase, merge, squash merge, cherry-pick, revert, sequencer, or bisect in a
  worktree is deferred and reported. The observer does not reconstruct a past operation from a
  post-commit callback.
- Dirty and detached worktrees are reported as attention items rather than auto-committed or
  auto-pushed. A failed discovery, state write, or push is visible in both the command result and
  durable status/log, then retried by the next scheduled run.
- Enforce the captain's approximately ten-minute cost floor with an atomically persisted
  per-push-target attempt timestamp written before a push attempt. A crashed lock holder releases
  its kernel lock automatically; no stale lease can permanently suppress later scans.

## Explicit non-goals

- No Git hook, commit delay, commit failure, force-push, default-branch push, merge, credential
  handling, external service, or new dependency.
- No attempt to push uncommitted work: it is made visible for a human/worker to commit safely.
- No automated enablement in this clone. Installation and removal of the optional launchd schedule
  remain explicit operator actions documented with the script.

## Hazard disposition from the abandoned hook report

| Hazard | Treatment |
|---|---|
| R1/R6/R7/R11/R12/R13 — past-operation reconstruction | Avoided structurally: polling checks only the current worktree state and never correlates a commit with a past operation. |
| R3/R8/R9 — competing/stale hook leases | Avoided structurally: one process holds a kernel-released advisory lock; busy/error states are reported, and a dead process leaves no held lock. |
| R4 — delayed observation-to-push race | Defended: rescan and operation/ref/remote recheck immediately before the explicit SHA push; the push is non-force and cannot rewrite a branch even if a concurrent change wins. |
| R2/R5/R10 — default/push-target ambiguity | Defended: resolve Git's effective push-remote precedence, reject multiple push URLs, and discover `HEAD` from that exact live push URL; uncertainty fails closed and is recorded. |
