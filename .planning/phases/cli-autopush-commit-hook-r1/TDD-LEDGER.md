# TDD ledger — Time-boxed rebase-safe post-commit push

Manual GSD TDD execution. The initial missing-hook failure is retained in
`traces/red-run.txt`; the same harness is the green evidence.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | No operation replay pushes | A two-commit real rebase can invoke a post-commit push once per replay. | The hook is invoked during both replays, sees Git's rebase marker through `--git-path`, and makes no push. |
| R2 | Default/detached refusal | `main` or a detached HEAD can reach `git push`. | Both commits succeed while their local commit remains absent from the local bare remote. |
| R3 | Per-branch ten-minute bound | A second feature commit can enqueue a second push immediately. | The first push succeeds; the second is skipped; an expired timestamp makes the next push carry all commits. |
| R4 | Linked-worktree state | A state path can be worktree-private when `.git` is assumed to be a directory or `core.hooksPath` is a worktree path. | The same branch state path resolved with the common Git directory and `git rev-parse --git-path` is identical from a linked worktree. |
| R5 | Never force and swallow rejection | A diverged remote can cause a force option or make the commit fail. | A real non-fast-forward remote rejects the exact non-force refspec, the remote remains unchanged, one local failure line is recorded, and the commit exits zero. |
| R6 | Detached execution | A slow push can hold the committing terminal open. | A temporary bare remote delays its actual receive; the commit returns before that delay while the remote update later completes. |
| R7 | Operator control | The hook is silently enabled or cannot be disabled. | Documentation leaves `core.hooksPath` opt-in and documents `PM_NO_AUTOPUSH=1`; the harness proves the variable skips pushing. |

## Red command

```sh
sh scripts/tests/post-commit-autopush.sh
```

The first run must fail because `.githooks/post-commit` does not yet exist. The
test is intentionally local-only: each case creates a temporary bare repository.

## Green commands

```sh
shellcheck -s sh .githooks/post-commit scripts/tests/post-commit-autopush.sh
sh scripts/tests/post-commit-autopush.sh
```
