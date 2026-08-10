# Code review — unpushed-work safety net r1

Command path: `scripts/gsd prompt code-review cli-unpushed-work-safety-net-r1`.
Execution: inline/manual because the task's canonical single-worker contract prohibits a spawned
reviewer role in this isolated worktree.

## Scope reviewed

- `scripts/unpushed-work-safety-net.py`
- `scripts/tests/unpushed-work-safety-net_test.py`
- `docs/unpushed-work-safety-net.md`, `docs/GUIDE.md`, and the Make target

## Result

No unresolved critical, warning, or informational findings.

## Findings fixed during the review pass

| ID | Severity | Finding | Disposition |
|---|---|---|---|
| R1 | warning | A generic `MERGE_MSG` could label a real conflicted cherry-pick as a merge. | Fixed by prioritizing `CHERRY_PICK_HEAD`, `REVERT_HEAD`, and other specific markers before the message-only fallback; real operation tests now assert each label. |
| R2 | warning | A push-target/default failure found during the final recheck could be reported but still permit a zero scan result. | Fixed by returning an explicit recheck-error signal and propagating it to the scan exit status. |
| R3 | warning | A durable event-log write failure was visible on stderr but could leave a successful process exit. | Fixed with reporter liveness state; a log failure now produces a nonzero result. |
| R4 | warning | A `Path` event field was not JSON serializable, turning detached-worktree reporting into a traceback. | Fixed by normalizing path fields before appending JSONL; the detached recovery test covers it. |
| R5 | info | The original tests proved ordinary divergence but not a real rejection after the final recheck. | Added a real pre-push race that advances the remote and asserts a non-fast-forward rejection preserves its SHA. |
| R6 | warning | A malformed push URL beginning with a dash could be interpreted as a Git option in the transport probes or the push invocation. | Fixed by terminating Git option parsing before every direct push URL; normal real-Git direct-URL tests pass. |
| R7 | info | The `remote_changed` final-recheck defer path lacked direct real-Git recovery proof. | Added a wrapper-assisted real remote ref update between the two `ls-remote` observations; it emits `remote_changed`, does not push, and a later scan recovers. |

## Safety review notes

- The only push construction is a pinned SHA refspec sent directly to the one live-validated push
  URL, with no force option or `+` prefix. The test wrapper refuses all force forms for every
  observer Git invocation, and a real remote rejects a non-fast-forward update.
- Push URL/default discovery never prints URLs or Git stderr, avoiding accidental credential
  exposure in events.
- The process lock is an `fcntl` advisory lock scoped to the shared common Git directory. It is
  released by the kernel when its holder exits, and contention is an explicit exit-75 event.
- No hook configuration is read or written by production code; the existing pre-commit hook stays
  an independent opt-in choice.
