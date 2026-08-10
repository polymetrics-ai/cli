# Discussion log — cli-autopush-commit-hook-r1

## GSD discussion execution

`scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` passed in this
worktree. The resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts were generated with `scripts/gsd prompt`.

This task is not a numbered phase in `.planning/ROADMAP.md`, Pi is not the active
runtime for this worker, and the task explicitly requires a single autonomous
worker. The repository's documented inline/manual GSD fallback therefore applies.
This file and the companion plan/TDD/verification artifacts are the lifecycle
record; no role or isolated-agent spawning is used.

## Fixed decisions

| Area | Decision | Source |
| --- | --- | --- |
| Enablement | Add a tracked hook only; do not set `core.hooksPath` or add an installer. | Task brief |
| Safety boundary | Skip all in-progress Git operations, detached HEAD, `main`, and a locally known remote default branch; a shared Git-path helper covers sequencer, squash, and merge-message states, while the detached child re-resolves it, resolves every effective push endpoint's live default, and only sends its scheduling OID if the branch still names it; never force-push. | Task brief + review follow-up |
| Operation cleanup | Snapshot Git-resolved operation state in `prepare-commit-msg` and consume it from `post-commit` by matching the commit parent. | Review follow-up |
| Rate limit | Store a per-branch timestamp under the shared Git hooks directory resolved with `git rev-parse --git-path`; an owner-backed atomic lease makes one attempt per 600 seconds across linked worktrees and reclaims only dead stale owners. | Task brief + review follow-up |
| Commit latency | Write the timestamp locally, then launch the push with standard input/output detached. The hook itself always exits zero. | Task brief |
| Failure behavior | A rejected or failed asynchronous push is swallowed and recorded as one short local log line. | Task brief |
| Testing | Use a POSIX-shell harness with temporary local bare remotes. It must run a real two-commit rebase, clean and manually completed operation states, delayed children during and after an operation, stale and pushurl-specific remote defaults, concurrent linked worktree leasing plus dead-lease recovery, and a real non-fast-forward rejection. | Task brief + review follow-up |
| Documentation | Explain opt-in setup and `PM_NO_AUTOPUSH=1` next to the existing hook instructions. | Task brief |

No product ambiguity remains, no credentials are used, and the tests use only local Git repositories.
