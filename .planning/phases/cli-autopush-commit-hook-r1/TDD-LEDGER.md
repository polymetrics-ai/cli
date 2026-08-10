# TDD ledger — Time-boxed rebase-safe post-commit push

Manual GSD TDD execution. The initial missing-hook failure is retained in
`traces/red-run.txt`; the same harness is the green evidence.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | No operation replay pushes | A two-commit real rebase or a manually completed merge, squash, cherry-pick, or revert can reach `post-commit` only after Git clears its operation marker. | A shared helper covers Git-resolved operation paths, including `SQUASH_MSG` and `MERGE_MSG`; the prepare hook records a durable pre-transition marker before cleanup, and every real operation case makes no push. |
| R2 | Default/detached refusal | `main`, a detached HEAD, or the live default at any effective push endpoint after its local tracking symref becomes stale can reach `git push`. | Both commits succeed while their local commit remains absent from the local bare remote; the detached child resolves every effective push URL's live `HEAD` and refuses the stale-default branch. |
| R3 | Per-branch ten-minute bound | Concurrent post-commit processes can each see an absent or expired timestamp and enqueue a push. | A shared `git update-ref` compare-and-swap permits one scheduled receiver call across forced linked worktrees from an expired timestamp; ordinary window and expiry cases still prove skip and catch-up. |
| R4 | Linked-worktree state | A state path can be worktree-private when `.git` is assumed to be a directory or `core.hooksPath` is a worktree path. | Rate state resolves from the common Git directory and is identical from a linked worktree; operation transition records resolve from that worktree's actual Git directory. |
| R5 | Never force and swallow rejection | A diverged remote can cause a force option or make the commit fail. | A real non-fast-forward remote rejects the exact non-force refspec, the remote remains unchanged, one local failure line is recorded, and the commit exits zero. |
| R6 | Detached execution | A slow push can hold the committing terminal open. | A temporary bare remote delays its actual receive; the commit returns before that delay while the remote update later completes. |
| R7 | Operator control | The hook is silently enabled or cannot be disabled. | Documentation leaves `core.hooksPath` opt-in and documents `PM_NO_AUTOPUSH=1`; the harness proves the variable skips pushing. |
| R8 | Delayed operation refusal | A delayed child can push while a later Git operation is active or can push a merge, cherry-pick, revert, or rebase tip that replaced the commit which scheduled it. | The child reuses the shared Git-path helper at its send boundary, captures the scheduling HEAD, requires the branch still equal that OID, and pushes the captured OID rather than a live branch ref. |
| R9 | Push-target default refusal | A fetch URL can have a non-default branch while a configured `pushurl` defaults to the committed branch. | Every effective push URL resolves its own live default before a single `git push`; a two-pushurl local case refuses both targets when either is default. |
| R10 | Configured push destination | A branch can fetch from `upstream` but explicitly configure a branch-specific or global push remote. | The hook selects `branch.<name>.pushRemote`, then `remote.pushDefault`, then `branch.<name>.remote`; local bare remotes prove each stage exclusively receives its commit. |
| R11 | Amended operation refusal | An amended squash or clean no-commit completion replaces the old tip, so the old tip is not the new commit's parent. | A transition record includes the old tip's full parent list, so amended squash, cherry-pick, and revert cases match and record no push. |
| R12 | Concurrent operation ownership | A delayed normal post hook can observe and consume another commit's operation marker after that operation updates the same branch. | Durable transition records are matched only to the candidate OID by direct parent or full amended-parent-list equality in both parent and child; an interleaved ordinary/squash case records no operation push and a later ordinary commit catches up. |

## Red command

```sh
sh scripts/tests/post-commit-autopush.sh
```

The first run must fail because `.githooks/post-commit` does not yet exist. The
test is intentionally local-only: each case creates a temporary bare repository.

## Green commands

```sh
shellcheck -s sh .githooks/pm-autopush-operation-state .githooks/prepare-commit-msg .githooks/post-commit scripts/tests/post-commit-autopush.sh
sh scripts/tests/post-commit-autopush.sh
```
