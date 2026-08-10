# Review — Time-boxed rebase-safe post-commit push

## Manual GSD code-review fallback

The generated `code-review` prompt was resolved, but the task requires a
single worker and the available runtime cannot provide an isolated review
agent. A focused manual review was performed after the executable harness and
shellcheck passed.

## Scope reviewed

- `.githooks/post-commit`
- `.githooks/prepare-commit-msg`
- `.githooks/pm-autopush-operation-state`
- `scripts/tests/post-commit-autopush.sh`
- `docs/GUIDE.md`
- the companion GSD evidence

## Follow-up findings

- R1 was legitimate: Git can remove a manual merge, cherry-pick, or revert
  marker before `post-commit`. A tracked `prepare-commit-msg` companion now
  records Git-resolved operation transitions in the worktree Git directory, and
  parent and child match their candidate commit against those records.
- R2 was legitimate: a local remote `HEAD` symref can be absent or stale. The
  detached child resolves the remote's live symbolic `HEAD`, logs an unavailable
  or matching default branch, and returns without push.
- R3 was legitimate: the old read/check/write sequence was not a lease. A
  common-Git-dir sibling directory now atomically protects the rate decision and
  timestamp replacement.
- R4 was legitimate: a delayed detached child could observe and send a later
  operation-generated branch tip, or send while a subsequent operation was active.
  It now re-resolves Git operation paths, captures the scheduling HEAD, requires
  the branch to remain at that OID before sending, and uses that captured OID as
  the non-force push source.
- R5 was legitimate: `git push <remote>` can use configured `pushurl` targets
  that differ from the fetch URL checked by `ls-remote`. The child now resolves
  and checks the live default of every effective push URL before it invokes the
  named remote.
- R6 and R7 were legitimate: clean squash, cherry-pick, and revert paths can
  rely on `SQUASH_MSG` or `MERGE_MSG` after their `*_HEAD` markers are absent.
  A shared Git-path helper now covers those paths in prepare, parent, and child
  checks, preventing future operation-list drift.
- R8 was legitimate: a process killed after acquiring the old directory lock
  could block that branch permanently. The hand-rolled lease is removed; the
  shared timestamp is now recorded with Git's atomic ref update.
- R9 was legitimate: stale hard-link recovery could unlink a newer live lease
  after an earlier stale observation. A compare-and-swap update of the shared
  timestamp ref makes the rate decision linearizable without stale takeover.
- R10 was legitimate: the parent chose a tracking remote before Git's explicit
  push destination settings. It now resolves branch-specific, global, and
  tracking remotes in Git's push precedence order.
- R11 was legitimate: an amended operation completion replaces the old tip, so
  it cannot be recognized from the new commit's direct parent. Each transition
  record includes the old tip's full parent list, which identifies amended
  completions without relying on that direct-parent relationship.
- R12 was legitimate: a delayed ordinary post hook can observe another commit's
  worktree-wide marker and consume it first. Transition records are durable and
  are matched only against a candidate's direct parent or full amended-parent
  list in both the parent and detached child paths.

The focused executable harness covers each follow-up, including clean and
amended squash, cherry-pick, and revert paths, an interleaved ordinary/squash
post-hook race, concurrent expired-rate compare-and-swap, delayed children during
and after a manual merge, mixed safe/default push URLs, and branch-specific/global/
tracking remote selection, alongside the original non-force rejection and
detached-push behavior.
