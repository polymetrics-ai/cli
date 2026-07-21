# Context: Issue #471 Pi AgentSession Shepherd

Issue: https://github.com/polymetrics-ai/cli/issues/471
Branch: `feat/471-pi-agent-session-shepherd`
Base: `origin/main` at `74ab381eb8236305170ffd44d5aed74f8d0d2936`
Canary target: draft PR #438 at exact head `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`

## Decision

Implement an experimental interactive Shepherd with Pi 0.80.6 public `AgentSession` APIs. It may
run independent embedded read-only sessions, collect bounded structured evidence, compute a
diagnostic rating, and enforce deterministic hard gates. It is not a process-isolation or reboot
durability boundary and it does not replace the Go Shepherd without a later human decision.

Issue #470 remains the existing Go-Shepherd/tmux control-surface contract and is blocked by #389.
Issue #471 is standalone so this experiment can target `main` and be canaried against PR #438
without silently rewriting the parent/sub-issue topology under #372.

## Pi SDK facts

- Installed Pi version: `0.80.6`.
- Public package exports include `createAgentSession`, `SessionManager.inMemory`,
  `SettingsManager.inMemory`, `DefaultResourceLoader`, `prompt`, `subscribe`, `abort`,
  `waitForIdle`, and `dispose`.
- A nested resource loader must use `noExtensions: true` and be explicitly reloaded.
- Child cancellation is cooperative through `AgentSession.abort()`; there is no per-child process
  kill boundary.
- Every embedded session shares the parent Node process, event loop, heap, environment, and crash
  domain. Tool calls may still start operating-system processes.
- The initial live canary is read-only. Mutation support remains fail-closed until an isolated
  worktree and declared write scope are supplied and separately validated.

## Workflow fallback

`scripts/gsd doctor` passes on `origin/main`, but that branch does not expose the later
`programming-loop` command or `scripts/programming-loop.mjs` / `scripts/tdd-gate.mjs` helpers.
Both attempted adapter commands returned `unknown GSD command`. The repository-authorized manual
GSD lifecycle is therefore active, with RED, GREEN, refactor, verification, trace, and run-state
evidence maintained in this phase directory.
