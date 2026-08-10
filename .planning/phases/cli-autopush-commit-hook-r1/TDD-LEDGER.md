# TDD ledger — Time-boxed rebase-safe post-commit push

Manual GSD TDD execution. The initial missing-hook failure is retained in
`traces/red-run.txt`; the same harness is the green evidence.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | No operation replay pushes | A two-commit real rebase or a manually completed merge, squash, cherry-pick, or revert can reach `post-commit` only after Git clears its operation marker. | A shared helper covers Git-resolved operation paths, including `SQUASH_MSG` and `MERGE_MSG`; the prepare hook snapshots state before cleanup, and every real operation case makes no push. |
| R2 | Default/detached refusal | `main`, a detached HEAD, or the live default at any effective push endpoint after its local tracking symref becomes stale can reach `git push`. | Both commits succeed while their local commit remains absent from the local bare remote; the detached child resolves every effective push URL's live `HEAD` and refuses the stale-default branch. |
| R3 | Per-branch ten-minute bound | Concurrent post-commit processes can each see an absent timestamp and enqueue a push, while an untrappable exit can strand a lease forever. | An owner-backed atomic shared lease permits one scheduled receiver call across forced linked worktrees and safely reclaims a dead owner after the rate window; ordinary window and expiry cases still prove skip and catch-up. |
| R4 | Linked-worktree state | A state path can be worktree-private when `.git` is assumed to be a directory or `core.hooksPath` is a worktree path. | Rate state resolves from the common Git directory and is identical from a linked worktree; operation snapshots resolve from that worktree's actual Git directory. |
| R5 | Never force and swallow rejection | A diverged remote can cause a force option or make the commit fail. | A real non-fast-forward remote rejects the exact non-force refspec, the remote remains unchanged, one local failure line is recorded, and the commit exits zero. |
| R6 | Detached execution | A slow push can hold the committing terminal open. | A temporary bare remote delays its actual receive; the commit returns before that delay while the remote update later completes. |
| R7 | Operator control | The hook is silently enabled or cannot be disabled. | Documentation leaves `core.hooksPath` opt-in and documents `PM_NO_AUTOPUSH=1`; the harness proves the variable skips pushing. |
| R8 | Delayed operation refusal | A delayed child can push while a later Git operation is active or can push a merge, cherry-pick, revert, or rebase tip that replaced the commit which scheduled it. | The child reuses the shared Git-path helper at its send boundary, captures the scheduling HEAD, requires the branch still equal that OID, and pushes the captured OID rather than a live branch ref. |
| R9 | Push-target default refusal | A fetch URL can have a non-default branch while a configured `pushurl` defaults to the committed branch. | Every effective push URL resolves its own live default before a single `git push`; a two-pushurl local case refuses both targets when either is default. |

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
