# Unpushed-work safety net

`scripts/unpushed-work-safety-net.py` is an opt-in, out-of-band observer for
feature-branch work that has been committed locally but not published. It is
deliberately separate from Git hooks: it does **not** set `core.hooksPath`, does
not invoke `.githooks/pre-commit`, and never delays or fails a commit.

The observer is intended for worker worktrees where an unnoticed local-only
branch is more costly than one bounded, automatic feature-branch push. It does
not commit, stash, reset, merge, rebase, force-push, or push a default branch.

## Explicit opt-in

From any worktree in the repository, enable the local configuration and create
the observer state directory:

```bash
python3 scripts/unpushed-work-safety-net.py enable
```

The configured minimum interval is 600 seconds. `enable` refuses a shorter
interval. It performs no scan or push itself.

For macOS, generate a repository-unique launchd job and explicitly load it:

```bash
LABEL="$(python3 scripts/unpushed-work-safety-net.py launchd-label)"
PLIST="$HOME/Library/LaunchAgents/$LABEL.plist"
python3 scripts/unpushed-work-safety-net.py launchd-plist > "$PLIST"
launchctl bootstrap "gui/$(id -u)" "$PLIST"
```

The generated job runs `run` every 600 seconds and writes launchd stdout/stderr
beside its durable state under the repository's Git common directory. Review
the generated plist before loading it; it is an operator-owned schedule, not a
repository-side automatic installation.

For another scheduler, invoke this one-shot command no more frequently than
every ten minutes:

```bash
python3 scripts/unpushed-work-safety-net.py run
```

## What a scan can do

For every linked worktree, the observer reports its outcome to stdout and to
`$(git rev-parse --git-common-dir)/unpushed-work-safety-net/events.jsonl`.
It only pushes when all of the following are true:

- the worktree is clean, attached to a branch, and not in a rebase, merge,
  squash merge, cherry-pick, revert, sequencer, or bisect;
- the branch's effective push remote is unambiguous: Git's
  `branch.<name>.pushRemote`, then `remote.pushDefault`, then tracking-remote
  precedence resolves to one remote with exactly one push URL;
- the default branch is discovered live from that **push URL**, not a local
  `origin/HEAD` cache or the remote's fetch URL, and the branch is not it;
- the remote branch is absent or is a known ancestor of the local branch; and
- the previous attempted push to that remote/branch was at least 600 seconds
  ago.

The final check re-reads operation state, worktree cleanliness, the branch SHA,
the effective push destination, its default branch, and the remote branch
before it sends anything. The push is from the shared Git database using a
pinned `<sha>:refs/heads/<branch>` refspec with no `+` and no force option. A
concurrent change can therefore defer or reject the push, but cannot cause a
remote rewrite.

Dirty work, a detached `HEAD`, a default branch, an active operation, a rate
limit, or divergence is not treated as a successful push. It is reported for a
worker or human to resolve. The tool never tries to publish uncommitted files.

## Health and explicit opt-out

Check whether the periodic observer has recently completed a healthy scan:

```bash
python3 scripts/unpushed-work-safety-net.py status
```

`status` returns a visible error for an invalid state file, a failed last scan,
or a stale heartbeat. `run` also exits nonzero for an unknown remote/default,
state failure, rejected push, or unavailable remote. These conditions stay in
the event log and are retried by the next scheduled run; a failure cannot look
like a clean push.

To opt out, first disable scans at the repository level, then unload and remove
the explicit scheduler entry:

```bash
python3 scripts/unpushed-work-safety-net.py disable
launchctl bootout "gui/$(id -u)" "$PLIST"
rm -f "$PLIST"
```

`disable` does not touch existing Git hooks. If a scheduler is still present,
its next run reports `event=not_enabled` and performs no push, making an
incomplete removal visible.

## Why this shape is safe

The abandoned post-commit design had to infer which *past* commit belonged to
an operation after Git had already removed its markers. This observer instead
examines only a current, settled worktree and leaves active work alone.

| Recorded hook hazard | Treatment |
|---|---|
| Rebase/merge/cherry-pick/revert/squash cleanup and cross-commit marker theft | Avoided structurally: no post-commit callback, synthetic marker, or past-operation inference exists. |
| Detached child observes a later ref state | Defended: scan snapshots a SHA and immediately rechecks current state before a pinned, non-force push from the common Git database. |
| Competing hooks and stale leases | Avoided structurally: the observer uses a non-blocking kernel advisory lock. A dead holder loses the lock automatically; a live holder produces `event=observer_busy`. |
| Wrong remote, fetch/push URL mismatch, and unknown default branch | Defended: resolve the effective push remote, reject multiple push URLs, and discover `HEAD` from the actual push URL on every candidate push. |

The integration harness exercises real bare remotes, linked worktrees, dirty
work recovery, a genuine conflicted two-commit rebase, diverged remote history,
push URL precedence, a corrupt state file, a live kernel lock, and the rate
floor:

```bash
python3 scripts/tests/unpushed-work-safety-net_test.py
```
